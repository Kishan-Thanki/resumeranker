"""arq worker entrypoint.

Real pipeline:
  queued → processing → (extract JD → extract resume → score) → completed
On LLMError → failed with the error message.
Any other exception → failed with "Internal error".
"""

import logging
import uuid
from datetime import UTC, datetime
from typing import Any, ClassVar

from arq.connections import RedisSettings
from sqlalchemy import select
from yarl import URL

from app.config import settings
from app.db import async_session_factory
from app.domain.tech import TechDomain
from app.models.analysis import Analysis
from app.services import llm_cost, llm_service
from app.services.llm_service import LLMError

logger = logging.getLogger(__name__)


def _redis_settings_from_url(url: str) -> RedisSettings:
    parsed = URL(url)
    return RedisSettings(
        host=parsed.host or "redis",
        port=parsed.port or 6379,
        database=int(parsed.path.lstrip("/") or "0"),
    )


async def _get_analysis(analysis_id: uuid.UUID) -> Analysis | None:
    async with async_session_factory() as db:
        result = await db.execute(select(Analysis).where(Analysis.id == analysis_id))
        return result.scalar_one_or_none()


async def _set_status(
    analysis_id: uuid.UUID,
    status_value: str,
    *,
    sections: list[dict[str, Any]] | None = None,
    error_message: str | None = None,
) -> None:
    async with async_session_factory() as db:
        result = await db.execute(select(Analysis).where(Analysis.id == analysis_id))
        row = result.scalar_one_or_none()
        if row is None:
            logger.warning("analysis %s missing — skipping", analysis_id)
            return
        row.status = status_value
        if sections is not None:
            row.sections = sections
        if error_message is not None:
            row.error_message = error_message
        row.updated_at = datetime.now(UTC)
        await db.commit()


async def run_analysis(ctx: dict[str, object], analysis_id: str) -> None:
    aid = uuid.UUID(analysis_id)
    logger.info("run_analysis start: %s", aid)

    row = await _get_analysis(aid)
    if row is None:
        return

    domain = TechDomain()

    try:
        await _set_status(aid, "processing")

        requirements = await llm_service.extract_jd_requirements(
            row.jd_text, domain, analysis_id=str(aid)
        )
        claims = await llm_service.extract_resume_claims(
            row.resume_text, domain, analysis_id=str(aid)
        )
        sections = await llm_service.score_requirements_against_claims(
            requirements,
            claims,
            domain,
            jd_text=row.jd_text,
            resume_text=row.resume_text,
            analysis_id=str(aid),
        )

        # Pydantic SectionScore[] → JSONB-compatible list[dict].
        sections_json = [s.model_dump(mode="json") for s in sections]
        await _set_status(aid, "completed", sections=sections_json)

        totals = await llm_cost.get_totals(str(aid))
        logger.info(
            "run_analysis complete: %s sections=%d tokens=%d/%d cost=$%.4f calls=%d",
            aid,
            len(sections_json),
            totals["prompt_tokens"],
            totals["completion_tokens"],
            totals["cost_usd"],
            totals["call_count"],
        )

    except LLMError as exc:
        logger.warning("run_analysis %s LLM failure: %s", aid, exc)
        await _set_status(aid, "failed", error_message=str(exc))
    except Exception:
        logger.exception("run_analysis %s crashed", aid)
        await _set_status(aid, "failed", error_message="Internal error")


class WorkerSettings:
    # arq inspects these as class-level attributes (not instance vars), so we
    # tag them ClassVar to satisfy ruff RUF012.
    functions: ClassVar[list[object]] = [run_analysis]
    redis_settings: ClassVar[RedisSettings] = _redis_settings_from_url(settings.redis_url)
    queue_name: ClassVar[str] = "resume_ranker"
    max_jobs: ClassVar[int] = 4
    job_timeout: ClassVar[int] = 120
    keep_result: ClassVar[int] = 0

"""Analysis service — no FastAPI imports.

DB row ↔ Pydantic schema conversion + ownership-scoped queries.
"""

import uuid

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.analysis import Analysis
from app.schemas.analysis import AnalysisResult, SectionScore


def to_schema(row: Analysis) -> AnalysisResult:
    """DB row → Pydantic. sections JSONB is already a list of dicts."""
    sections = [SectionScore.model_validate(s) for s in (row.sections or [])]
    return AnalysisResult(
        id=str(row.id),
        created_at=row.created_at,
        jd_title=row.jd_title,
        resume_name=row.resume_name,
        status=row.status,
        sections=sections,
        error_message=row.error_message,
    )


async def list_for_user(db: AsyncSession, user_id: uuid.UUID) -> list[AnalysisResult]:
    result = await db.execute(
        select(Analysis).where(Analysis.user_id == user_id).order_by(Analysis.created_at.desc())
    )
    return [to_schema(row) for row in result.scalars().all()]


async def get_for_user(
    db: AsyncSession, user_id: uuid.UUID, analysis_id: str
) -> AnalysisResult | None:
    try:
        analysis_uuid = uuid.UUID(analysis_id)
    except ValueError:
        return None
    result = await db.execute(
        select(Analysis).where(Analysis.id == analysis_uuid, Analysis.user_id == user_id)
    )
    row = result.scalar_one_or_none()
    return to_schema(row) if row is not None else None


async def delete_for_user(db: AsyncSession, user_id: uuid.UUID, analysis_id: str) -> bool:
    """Hard-delete one analysis the user owns. Returns True if a row was
    removed, False if the id was malformed or didn't belong to the user.

    The `user_id` predicate is the ownership gate — a non-owner's delete
    matches zero rows and returns False, which the route surfaces as the
    same 404 a missing analysis gets (no existence leak). The row holds
    the encrypted jd_text / resume_text and the scores, so removing it
    discharges the "deleting an analysis removes its content" promise in
    the privacy policy.
    """
    try:
        analysis_uuid = uuid.UUID(analysis_id)
    except ValueError:
        return False
    result = await db.execute(
        select(Analysis).where(Analysis.id == analysis_uuid, Analysis.user_id == user_id)
    )
    row = result.scalar_one_or_none()
    if row is None:
        return False
    await db.delete(row)
    await db.commit()
    return True

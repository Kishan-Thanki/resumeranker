from typing import Annotated, Literal

import redis.asyncio as redis_async
from fastapi import (
    APIRouter,
    Depends,
    File,
    Form,
    HTTPException,
    Request,
    UploadFile,
    status,
)
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.db import get_db
from app.deps import get_current_user, get_redis
from app.models.user import User
from app.schemas.analysis import AnalysisResult
from app.services import analysis_create, analysis_service, audit_service
from app.services.job_queue import enqueue_analysis
from app.services.pdf_service import PdfExtractionError, extract_text_from_pdf
from app.services.rate_limit import RateLimitExceeded, check_and_increment

router = APIRouter(prefix="/analyses", tags=["analyses"])

MAX_JD_TEXT_CHARS = 50_000
PDF_MIME = "application/pdf"


@router.get("", response_model=list[AnalysisResult], response_model_by_alias=True)
async def list_analyses(
    db: Annotated[AsyncSession, Depends(get_db)],
    user: Annotated[User, Depends(get_current_user)],
) -> list[AnalysisResult]:
    return await analysis_service.list_for_user(db, user.id)


@router.get("/{analysis_id}", response_model=AnalysisResult, response_model_by_alias=True)
async def get_analysis(
    analysis_id: str,
    db: Annotated[AsyncSession, Depends(get_db)],
    user: Annotated[User, Depends(get_current_user)],
) -> AnalysisResult:
    result = await analysis_service.get_for_user(db, user.id, analysis_id)
    if result is None:
        # Don't leak existence to non-owners — same 404 as missing.
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="not found")
    return result


@router.delete("/{analysis_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_analysis(
    analysis_id: str,
    db: Annotated[AsyncSession, Depends(get_db)],
    user: Annotated[User, Depends(get_current_user)],
) -> None:
    # Granular right-to-erasure: removes this analysis' encrypted content
    # and scores. Ownership is enforced in the service via the user_id
    # predicate; a non-owner (or missing id) deletes zero rows and gets
    # the same 404 as a missing analysis — no existence leak.
    deleted = await analysis_service.delete_for_user(db, user.id, analysis_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="not found")


def _validate_pdf(file: UploadFile, label: str, max_mb: int) -> None:
    if file.content_type != PDF_MIME and not (file.filename or "").lower().endswith(".pdf"):
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"{label} must be a PDF",
        )
    if file.size is not None and file.size > max_mb * 1024 * 1024:
        raise HTTPException(
            status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
            detail=f"{label} exceeds {max_mb} MB limit",
        )


@router.post(
    "",
    response_model=AnalysisResult,
    response_model_by_alias=True,
    status_code=status.HTTP_201_CREATED,
)
async def create_analysis(
    request: Request,
    db: Annotated[AsyncSession, Depends(get_db)],
    redis: Annotated[redis_async.Redis, Depends(get_redis)],
    user: Annotated[User, Depends(get_current_user)],
    jd_input_type: Annotated[Literal["pdf", "text"], Form()],
    resume: Annotated[UploadFile, File()],
    jd_pdf: Annotated[UploadFile | None, File()] = None,
    jd_text: Annotated[str | None, Form()] = None,
) -> AnalysisResult:
    client_ip = request.client.host if request.client else "unknown"
    user_agent = request.headers.get("user-agent")

    # Service-wide daily cap — checked FIRST so a globally-exhausted day
    # doesn't burn the user's per-user counter. Returns 503 because this
    # is a service capacity issue, not a per-user abuse signal. The
    # counter is on a 24h rolling window from first use of the day.
    try:
        await check_and_increment(
            redis,
            "rl:analyses:global:daily",
            max_count=settings.max_analyses_per_day_global,
            window_seconds=24 * 3600,
        )
    except RateLimitExceeded as exc:
        await audit_service.log_event(
            audit_service.EventType.RATELIMIT_HIT,
            user_id=user.id,
            email=user.email,
            ip_address=client_ip,
            user_agent=user_agent,
            details={"scope": "analyses.global_daily", "key": str(exc)},
        )
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="daily service capacity reached — please try again tomorrow",
        ) from exc

    # Per-user daily quota
    try:
        await check_and_increment(
            redis,
            f"rl:analyses:user:{user.id}:daily",
            max_count=settings.max_resumes_per_user_per_day,
            window_seconds=24 * 3600,
        )
    except RateLimitExceeded as exc:
        await audit_service.log_event(
            audit_service.EventType.RATELIMIT_HIT,
            user_id=user.id,
            email=user.email,
            ip_address=client_ip,
            user_agent=user_agent,
            details={"scope": "analyses.per_user_daily", "key": str(exc)},
        )
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail=f"daily limit of {settings.max_resumes_per_user_per_day} analyses reached",
        ) from exc

    _validate_pdf(resume, "resume", settings.max_file_size_mb)
    resume_bytes = await resume.read()
    try:
        resume_text = extract_text_from_pdf(resume_bytes)
    except PdfExtractionError as exc:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=str(exc),
        ) from exc

    if jd_input_type == "pdf":
        if jd_pdf is None:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="jd_pdf is required when jd_input_type=pdf",
            )
        _validate_pdf(jd_pdf, "jd_pdf", settings.max_file_size_mb)
        jd_bytes = await jd_pdf.read()
        try:
            extracted_jd = extract_text_from_pdf(jd_bytes)
        except PdfExtractionError as exc:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail=str(exc),
            ) from exc
    else:
        if not jd_text or not jd_text.strip():
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="jd_text is required when jd_input_type=text",
            )
        if len(jd_text) > MAX_JD_TEXT_CHARS:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail=f"jd_text exceeds {MAX_JD_TEXT_CHARS} chars",
            )
        extracted_jd = jd_text

    analysis = await analysis_create.insert_queued_analysis(
        db=db,
        user=user,
        jd_text=extracted_jd,
        resume_text=resume_text,
        resume_name=resume.filename or "resume.pdf",
    )
    await enqueue_analysis(str(analysis.id))
    return analysis_service.to_schema(analysis)

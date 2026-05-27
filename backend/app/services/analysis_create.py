"""Create-flow logic for new analyses. Pure service, no FastAPI imports.

Splits the validated text inputs into a persisted `queued` row that the
worker will pick up. Phase 6 only inserts; Phase 7 wires the enqueue.
"""

import uuid

from sqlalchemy.ext.asyncio import AsyncSession

from app.models.analysis import Analysis
from app.models.user import User


def derive_jd_title(jd_text: str, fallback: str = "Untitled job description") -> str:
    for line in jd_text.splitlines():
        stripped = line.strip()
        if stripped:
            return stripped[:80]
    return fallback


async def insert_queued_analysis(
    db: AsyncSession,
    user: User,
    jd_text: str,
    resume_text: str,
    resume_name: str,
) -> Analysis:
    analysis = Analysis(
        id=uuid.uuid4(),
        user_id=user.id,
        status="queued",
        jd_title=derive_jd_title(jd_text),
        resume_name=resume_name,
        jd_text=jd_text,
        resume_text=resume_text,
        sections=[],
    )
    db.add(analysis)
    await db.commit()
    await db.refresh(analysis)
    return analysis

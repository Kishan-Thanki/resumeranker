import uuid
from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db import get_db
from app.deps import RequireRole
from app.models.user import User, UserRole
from app.schemas.auth import UserPublic
from app.schemas.iam import RoleUpdateBody

router = APIRouter(prefix="/iam", tags=["iam"])

# Require at least admin to list users
RequireAdmin = RequireRole([UserRole.ADMIN.value, UserRole.SUPERADMIN.value])
# Require superadmin to change roles
RequireSuperAdmin = RequireRole([UserRole.SUPERADMIN.value])


@router.get("/users", response_model=list[UserPublic])
async def list_users(
    user: Annotated[User, Depends(RequireAdmin)],
    db: Annotated[AsyncSession, Depends(get_db)],
) -> list[User]:
    """List all registered users. Admin-only."""
    result = await db.execute(select(User).order_by(User.created_at.desc()))
    return list(result.scalars().all())


@router.patch("/users/{user_id}/role", response_model=UserPublic)
async def update_user_role(
    user_id: uuid.UUID,
    body: RoleUpdateBody,
    current_user: Annotated[User, Depends(RequireSuperAdmin)],
    db: Annotated[AsyncSession, Depends(get_db)],
) -> User:
    """Update a user's role. Superadmin-only."""
    if user_id == current_user.id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot change your own role.",
        )

    target_user = await db.get(User, user_id)
    if not target_user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")

    target_user.role = body.role
    await db.commit()
    await db.refresh(target_user)
    return target_user

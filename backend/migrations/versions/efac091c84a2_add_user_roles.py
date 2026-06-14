"""add_user_roles

Revision ID: efac091c84a2
Revises: d62b68a21d7b
Create Date: 2026-05-29 10:37:50.651986

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'efac091c84a2'
down_revision: Union[str, None] = 'd62b68a21d7b'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('users', sa.Column('role', sa.VARCHAR(length=50), server_default='user', nullable=False))


def downgrade() -> None:
    op.drop_column('users', 'role')

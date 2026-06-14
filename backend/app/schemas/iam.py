from pydantic import BaseModel, ConfigDict
from app.models.user import UserRole

class RoleUpdateBody(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    role: UserRole

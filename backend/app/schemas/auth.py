import uuid

from pydantic import BaseModel, ConfigDict, EmailStr, Field


class RequestLinkBody(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    email: EmailStr
    # Click-wrap consent: the policy version string the user just saw and
    # checked the box for. Required (not Optional) so the request fails
    # validation if the frontend forgets to send it — that's the desired
    # behavior, the gate is mandatory.
    accepted_policy_version: str = Field(
        min_length=1,
        max_length=64,
        validation_alias="acceptedPolicyVersion",
        serialization_alias="acceptedPolicyVersion",
    )


class RequestLinkResponse(BaseModel):
    ok: bool = True


class VerifyBody(BaseModel):
    token: str = Field(min_length=1, max_length=512)


from app.models.user import UserRole


class UserPublic(BaseModel):
    model_config = ConfigDict(from_attributes=True, populate_by_name=True)

    id: uuid.UUID
    email: EmailStr
    role: UserRole
    # Exposed so the frontend can compare against its current build's
    # CURRENT_POLICY_VERSION and show a re-acceptance modal when the user
    # is on an older one. Nullable for backfilled pre-policy users.
    accepted_policy_version: str | None = Field(
        default=None,
        serialization_alias="acceptedPolicyVersion",
    )


class VerifyResponse(BaseModel):
    """Successful magic-link verify response.

    Note: the session token itself is set on an HttpOnly cookie via
    `Set-Cookie` and is deliberately NOT returned in this body — keeping
    it out of JS-readable memory is the entire reason for moving off
    localStorage. The body just confirms which user is now signed in.
    """

    model_config = ConfigDict(populate_by_name=True)

    user: UserPublic


class MeResponse(UserPublic):
    pass


class AcceptPolicyBody(BaseModel):
    """Body for `POST /me/accept-policy` — re-acceptance flow.

    Used when the user's stored `accepted_policy_version` is older than
    the frontend's `CURRENT_POLICY_VERSION`. The user clicks "I agree"
    on a modal; the frontend POSTs the new version here.
    """

    model_config = ConfigDict(populate_by_name=True)

    accepted_policy_version: str = Field(
        min_length=1,
        max_length=64,
        validation_alias="acceptedPolicyVersion",
        serialization_alias="acceptedPolicyVersion",
    )

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    database_url: str = "postgresql+asyncpg://resumeranker:resumeranker@postgres:5432/resumeranker"
    redis_url: str = "redis://redis:6379/0"

    # Single provider-agnostic LLM credential. The provider is determined
    # by `llm_model` (e.g. `gemini/gemini-2.5-flash` → Google AI Studio,
    # `anthropic/claude-haiku-4-5` → Anthropic). `llm_service.py` passes
    # this value explicitly to litellm via the `api_key` kwarg so we don't
    # depend on litellm's per-provider env-var auto-discovery.
    llm_api_key: str = "replace-me"
    llm_model: str = "gemini/gemini-2.5-flash"

    resend_api_key: str = "replace-me"
    resend_from_email: str = "login@yourdomain.com"

    # Inbox that /contact form submissions are forwarded to. Never exposed
    # in any user-facing response — purely a backend routing target. Set
    # to your real address (e.g. a personal Gmail) in `.env`. Falls back
    # to stub mode (logs the message to stdout) when unset / replace-me.
    contact_to_email: str = "replace-me"

    app_base_url: str = "http://localhost:5173"
    environment: str = "development"

    max_resumes_per_user_per_day: int = 10
    max_file_size_mb: int = 10

    # Service-wide daily cap on analyses, summed across all users. Lower
    # than the per-user cap × any reasonable user count — its job is to
    # keep us comfortably under the LLM provider's free tier so a single
    # busy visitor can't burn through it. Gemini's free tier is ~1M
    # tokens/day; one analysis is ~30-100K tokens (3 LLM calls), so 25
    # leaves headroom. Bump cautiously based on observed token usage.
    max_analyses_per_day_global: int = 25

    # Current ToS+Privacy version users must accept. Date of the last
    # material change to either policy, formatted YYYY-MM-DD. Bump this
    # value whenever you update /terms or /privacy in a way that requires
    # re-acceptance. The frontend hardcodes the same string in
    # `src/lib/policy.ts` — keep them in sync.
    current_policy_version: str = "2026-05-19"

    # Symmetric key used to encrypt resume_text and jd_text columns at the
    # application layer via Fernet. 32 random bytes, urlsafe-base64. Generate
    # with `python -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"`.
    # Rotating this value invalidates every existing analysis row — handle
    # rotation with a dedicated migration that re-encrypts with both keys.
    # Stub fallback ("replace-me") disables encryption so local/CI environments
    # without a key still boot; production MUST set a real value.
    resume_encryption_key: str = "replace-me"

    @property
    def encryption_stub_mode(self) -> bool:
        """True when no real encryption key is configured. Columns store plaintext."""
        return self.resume_encryption_key.strip() in {"", "replace-me", "your-key-here"}

    @property
    def llm_stub_mode(self) -> bool:
        """True when no real API key is configured. Pipeline returns canned fixtures."""
        return self.llm_api_key.strip() in {"", "replace-me", "your-key-here"}

    @property
    def email_stub_mode(self) -> bool:
        """True when no real Resend key is configured. Magic links print to stdout."""
        return self.resend_api_key.strip() in {"", "replace-me", "your-key-here"}


settings = Settings()

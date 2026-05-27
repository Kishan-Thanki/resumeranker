# Rules for AI assistant on this backend

Follow these rules when editing this repo. They come from the v1 backend
brief; if `BACKEND_BRIEF.md` is present at this directory it is the source
of truth, otherwise the brief lives in the conversation that created this
scaffold.

1. **No untyped code.** `mypy --strict` must pass. No `Any` without an
   explicit reason in a comment.
2. **Services have no FastAPI imports.** Routes call services; services
   don't know about HTTP.
3. **No file deletion** without explicit permission. If a file becomes
   obsolete, leave it and ask.
4. **All DB access is async** via SQLAlchemy async session. No sync calls
   anywhere in request or worker paths.
5. **Pydantic for every boundary.** Request bodies, response bodies, LLM
   responses, env config. No free dicts crossing boundaries.
6. **Migrations are immutable once committed.** Never edit an applied
   migration; add a new one.
7. **Never log secrets.** No tokens, no API keys, no full session tokens
   in logs. Log token IDs / prefixes only.
8. **Never log full PII unprompted.** Email addresses can be logged.
   Resume contents cannot.
9. **Routes are thin.** Validation + dependency injection + one service
   call. Business logic lives in services.
10. **Worker is the only place LLM calls happen.** Routes never call
    litellm directly.
11. **All LLM responses go through `instructor` + a Pydantic model.** No
    free-text outputs.
12. **Prompt changes bump `PROMPT_VERSION`.** This invalidates caches and
    triggers eval re-runs.
13. **Docker-first workflow.** Never run `uv` or `python` on the host. All
    commands through `docker compose exec`.
14. **Test for the contract, not the implementation.** Pydantic response
    shapes are tested; internal service signatures aren't.
15. **One commit per working-order step** when working with AI assistants.
    Easier to revert, easier to review.

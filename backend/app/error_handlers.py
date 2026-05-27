"""Generic exception handlers.

FastAPI's default error shape (`{"detail": ...}`) plus the auto-generated
422 body from Pydantic (`{"detail": [{"loc": [...], "msg": "...", "type": ...}]}`)
are recognizable fingerprints. We replace both with a stable shape that
gives no framework signal: `{"error": "<slug>"}`.

The slug is generic enough that an attacker can't infer route internals
from it. Detailed error info is still logged server-side for ops.
"""

from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable
from typing import Any

from fastapi import FastAPI, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse, Response
from starlette.exceptions import HTTPException as StarletteHTTPException

logger = logging.getLogger(__name__)


def _slug_for_status(code: int) -> str:
    if code == status.HTTP_400_BAD_REQUEST:
        return "bad_request"
    if code == status.HTTP_401_UNAUTHORIZED:
        return "unauthorized"
    if code == status.HTTP_403_FORBIDDEN:
        return "forbidden"
    if code == status.HTTP_404_NOT_FOUND:
        return "not_found"
    if code == status.HTTP_405_METHOD_NOT_ALLOWED:
        return "method_not_allowed"
    if code == status.HTTP_409_CONFLICT:
        return "conflict"
    if code == status.HTTP_413_REQUEST_ENTITY_TOO_LARGE:
        return "payload_too_large"
    if code == status.HTTP_422_UNPROCESSABLE_ENTITY:
        return "unprocessable"
    if code == status.HTTP_429_TOO_MANY_REQUESTS:
        return "rate_limited"
    if code >= 500:
        return "server_error"
    return "error"


def _generic_body(code: int, detail: Any = None) -> dict[str, str]:
    body: dict[str, str] = {"error": _slug_for_status(code)}
    # Only echo a textual detail back when it's a known short user-facing
    # message set by a route handler (i.e. a string, not Pydantic's
    # validation-error list). This keeps "rate-limited", "invalid link",
    # etc. visible to legitimate clients without leaking framework guts.
    if (
        isinstance(detail, str)
        and detail
        and detail.lower() not in {"not found", "method not allowed"}
    ):
        body["message"] = detail
    return body


async def _http_exception_handler(_request: Request, exc: Exception) -> JSONResponse:
    # Starlette dispatches the matched exception to this handler. We take
    # the broad `Exception` type so the function signature satisfies
    # Starlette's typestub; the dispatch table guarantees `exc` is the
    # registered subclass at runtime.
    status_code = getattr(exc, "status_code", status.HTTP_500_INTERNAL_SERVER_ERROR)
    detail = getattr(exc, "detail", None)
    return JSONResponse(
        status_code=status_code,
        content=_generic_body(status_code, detail),
    )


async def _validation_exception_handler(_request: Request, exc: Exception) -> JSONResponse:
    # Pydantic's default body leaks `loc`, `msg`, `type` arrays which are a
    # FastAPI signature. Server-side, keep the rich info in the log.
    errors_method = getattr(exc, "errors", None)
    if callable(errors_method):
        logger.info("validation error", extra={"errors": errors_method()})
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=_generic_body(status.HTTP_422_UNPROCESSABLE_ENTITY),
    )


async def _unhandled_exception_handler(_request: Request, exc: Exception) -> JSONResponse:
    # Log the full exception with traceback for ops; never leak the message
    # or stack to the client.
    logger.exception("unhandled exception", exc_info=exc)
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=_generic_body(status.HTTP_500_INTERNAL_SERVER_ERROR),
    )


def register(app: FastAPI) -> None:
    """Install all generic handlers on `app`. Call from main.

    The `type: ignore[arg-type]` lines below silence a Starlette typestub
    quirk: the stub for `add_exception_handler` is over-strict about the
    handler's parameter type when an overload is selected by the exception
    class. Our handlers correctly take the broader `Exception` type, which
    is structurally compatible at runtime but mypy can't see through the
    overload resolution.
    """
    # mypy's view of add_exception_handler's overloads narrows the handler
    # param to the registered exception subclass; our handlers correctly
    # take the broader `Exception`. Runtime is fine, types disagree.
    from typing import cast

    handler_cb = Callable[[Request, Exception], Awaitable[Response]]
    app.add_exception_handler(StarletteHTTPException, cast(handler_cb, _http_exception_handler))
    app.add_exception_handler(
        RequestValidationError, cast(handler_cb, _validation_exception_handler)
    )
    app.add_exception_handler(Exception, _unhandled_exception_handler)

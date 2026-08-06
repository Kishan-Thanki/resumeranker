import contextvars
import json
import logging
import os
import traceback
from datetime import datetime, timezone

request_id_var: contextvars.ContextVar[str] = contextvars.ContextVar(
    "request_id", default=""
)

SERVICE_NAME = os.environ["SERVICE_NAME"]
DEBUG = os.environ["DEBUG"]

TRUE_VALUES = frozenset({"true", "1", "yes", "on"})

RESERVED_LOG_RECORD_FIELDS = frozenset(logging.makeLogRecord({}).__dict__)


class JSONFormatter(logging.Formatter):
    """
    Custom logging formatter that outputs log records as JSON strings.

    Automatically includes standard metadata (timestamp, level, logger, service),
    request correlation IDs, exception tracebacks, and any custom attributes
    supplied through the `extra` parameter.
    """

    def format(self, record: logging.LogRecord) -> str:
        """
        Formats a LogRecord into a JSON string.
        """
        log_obj: dict[str, object] = {
            "timestamp": datetime.fromtimestamp(
                record.created, tz=timezone.utc
            ).isoformat().replace("+00:00", "Z"),
            "level": record.levelname,
            "message": record.getMessage(),
            "logger": record.name,
            "service": SERVICE_NAME,
        }

        request_id = request_id_var.get()
        if request_id:
            log_obj["request_id"] = request_id

        for key, value in record.__dict__.items():
            if key not in RESERVED_LOG_RECORD_FIELDS:
                log_obj[key] = value

        if record.exc_info:
            log_obj["exception"] = "".join(
                traceback.format_exception(*record.exc_info)
            )

        return json.dumps(
            log_obj,
            ensure_ascii=False,
            default=str,
        )


def setup_logger(name: str) -> logging.Logger:
    """
    Initializes and configures the application's logger.

    In development (DEBUG enabled), logs are emitted in a human-readable format.
    In all other environments, logs are emitted as structured JSON suitable for
    centralized log aggregation systems.
    """
    logger = logging.getLogger(name)

    if logger.hasHandlers():
        logger.handlers.clear()

    debug_mode = DEBUG.strip().lower() in TRUE_VALUES

    logger.setLevel(logging.DEBUG if debug_mode else logging.INFO)

    handler = logging.StreamHandler()

    if debug_mode:
        formatter = logging.Formatter(
            "%(asctime)s - %(levelname)s - %(name)s - %(message)s"
        )
    else:
        formatter = JSONFormatter()

    handler.setFormatter(formatter)

    logger.addHandler(handler)
    logger.propagate = False

    return logger


logger = setup_logger(SERVICE_NAME)
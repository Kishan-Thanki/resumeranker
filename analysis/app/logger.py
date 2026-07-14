import logging
import json
import os
import traceback
import contextvars
from datetime import datetime

request_id_var: contextvars.ContextVar[str] = contextvars.ContextVar("request_id", default="")

class JSONFormatter(logging.Formatter):
    """
    Custom logging formatter that outputs log records as JSON strings.
    Automatically includes standard fields like timestamp, level, and message,
    along with any extra contextual attributes passed to the logger.
    """
    def format(self, record):
        """
        Formats a LogRecord into a JSON string.
        Injects the current request_id from contextvars and captures exception tracebacks.
        """
        log_obj = {
            "timestamp": datetime.utcfromtimestamp(record.created).isoformat() + "Z",
            "level": record.levelname,
            "message": record.getMessage(),
            "logger": record.name,
        }
        
        req_id = request_id_var.get()
        if req_id:
            log_obj["request_id"] = req_id

        standard_attrs = {
            'args', 'asctime', 'created', 'exc_info', 'exc_text', 'filename',
            'funcName', 'id', 'levelname', 'levelno', 'lineno', 'module',
            'msecs', 'message', 'msg', 'name', 'pathname', 'process',
            'processName', 'relativeCreated', 'stack_info', 'thread', 'threadName',
            'taskName'
        }
        
        for key, value in record.__dict__.items():
            if key not in standard_attrs:
                log_obj[key] = value

        if record.exc_info:
            log_obj["exception"] = "".join(traceback.format_exception(*record.exc_info))
            
        return json.dumps(log_obj)

def setup_logger(name: str = "analysis") -> logging.Logger:
    """
    Initializes and configures the standard Python logger for the application.
    Sets the log level based on the environment (defaulting to INFO) and
    attaches the JSONFormatter to the stream handler.
    """
    logger = logging.getLogger(name)
    
    if logger.hasHandlers():
        logger.handlers.clear()

    debug_mode = os.environ.get("DEBUG", "False").lower() in ("true", "1", "yes")
    
    logger.setLevel(logging.DEBUG if debug_mode else logging.INFO)
    
    handler = logging.StreamHandler()
    
    if debug_mode:
        formatter = logging.Formatter(
            fmt="%(asctime)s - %(levelname)s - %(name)s - %(message)s"
        )
    else:
        formatter = JSONFormatter()
        
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.propagate = False
    
    return logger

logger = setup_logger()

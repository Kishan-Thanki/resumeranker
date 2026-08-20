FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1

WORKDIR /app

COPY pyproject.toml README.md ./
COPY src ./src

RUN pip install --upgrade pip setuptools wheel \
    && pip install .

EXPOSE 8501

CMD ["streamlit", "run", "src/analysis_dashboard/app.py", "--server.address=0.0.0.0", "--server.port=8501", "--server.maxUploadSize=10"]

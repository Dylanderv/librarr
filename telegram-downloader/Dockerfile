FROM python:3.12-slim

WORKDIR /app

COPY pyproject.toml ./
RUN pip install --no-cache-dir .

COPY app ./app

ENV TELEGRAM_SESSION_PATH=/data/telegram/session \
    TELEGRAM_DB_PATH=/data/telegram/queue.db \
    PORT=5051

VOLUME ["/data/telegram"]
EXPOSE 5051

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5051"]

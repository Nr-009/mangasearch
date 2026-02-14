FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ocr_server.py .

EXPOSE 5000

ENV WORKERS=4

CMD uvicorn ocr_server:app --host 0.0.0.0 --port 5000 --workers ${WORKERS}
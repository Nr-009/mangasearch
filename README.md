# MangaSearch

A local CLI tool that indexes your manga collection and lets you search for quotes across it.

```
mangasearch index /manga/Berserk
mangasearch search "I sacrifice"
→ Berserk Chapter 57, Page 14: "I sacrifice all of you..."
```

## Stack

| Layer           | Tech                         |
| --------------- | ---------------------------- |
| App             | Go + Cobra + Gin             |
| OCR             | Python + FastAPI + Tesseract |
| Queue           | Redis                        |
| Source of truth | PostgreSQL                   |
| Search          | Elasticsearch                |
| Infrastructure  | Docker Compose               |

## Quick Start

```bash
# 1. Start infrastructure
docker compose up -d

# 2. Start OCR server
cd python
pip install -r requirements.txt
uvicorn ocr_server:app --workers 4 --port 5000

# 3. Run the CLI (coming soon)
go run main.go index /path/to/manga
```

## Architecture

See `ARCHITECTURE.md` for full data flow diagrams and design decisions.

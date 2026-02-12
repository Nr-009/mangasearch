from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import easyocr
import os

app = FastAPI(title="MangaSearch OCR Server")

reader = easyocr.Reader(['en'], gpu=False)

class OCRRequest(BaseModel):
    path: str

class OCRResponse(BaseModel):
    text: str


@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/ocr", response_model=OCRResponse)
def extract_text(payload: OCRRequest):
    path = payload.path
    
    if not os.path.exists(path):
        raise HTTPException(status_code=404, detail=f"File not found: {path}")

    results = reader.readtext(path)

    fragments = []
    for (bbox, text, confidence) in results:
        text = text.strip()
        if text and confidence > 0.4:
            fragments.append(text)

    page_text = " ".join(fragments)
    print(f"processing: {path}")
    return OCRResponse(text=page_text)
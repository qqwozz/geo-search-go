from fastapi import FastAPI
from pydantic import BaseModel
from typing import Optional
from parser import parse_query

app = FastAPI(title="Geo Search NLP")


class ParseRequest(BaseModel):
    text: str
    city: str = "moscow"


class ParseResponse(BaseModel):
    category: str
    intent: str
    features: dict
    location: Optional[dict] = None
    radius: int
    radius_raw: str
    sort_by: str


@app.post("/parse", response_model=ParseResponse)
async def parse(req: ParseRequest):
    return parse_query(req.text, req.city)


@app.get("/health")
async def health():
    return {"status": "ok"}

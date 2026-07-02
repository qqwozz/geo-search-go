from typing import Optional
from dictionaries import CATEGORIES, FEATURES, INTENTS, DISTANCE
from location_parser import extract_metro, extract_street


def parse_query(text: str, city: str = "moscow") -> dict:
    lower = text.lower()

    category = "cafe"
    for cat, keywords in CATEGORIES.items():
        for kw in keywords:
            if kw in lower:
                category = cat
                break

    features = {}
    for feat, keywords in FEATURES.items():
        for kw in keywords:
            if kw in lower:
                features[feat] = True
                break

    intent = "default"
    for i, keywords in INTENTS.items():
        for kw in keywords:
            if kw in lower:
                intent = i
                break

    radius = 2000
    radius_raw = ""
    for word, dist in DISTANCE:
        if word in lower:
            radius = dist
            radius_raw = word
            break

    metro = extract_metro(text)
    street = extract_street(text)

    location = None
    if metro or street:
        location = {}
        if metro:
            location["metro"] = metro
        if street:
            location["street"] = street

    return {
        "category": category,
        "intent": intent,
        "features": features,
        "location": location,
        "radius": radius,
        "radius_raw": radius_raw,
        "sort_by": "relevance",
    }

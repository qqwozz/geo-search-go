#!/usr/bin/env python3
import json
import os
import sys

try:
    import overpy
except ImportError:
    print("Installing overpy...")
    os.system(f"{sys.executable} -m pip install overpy")
    import overpy

try:
    import psycopg2
except ImportError:
    print("Installing psycopg2-binary...")
    os.system(f"{sys.executable} -m pip install psycopg2-binary")
    import psycopg2


def load_config():
    config_path = os.path.join(os.path.dirname(__file__), "ingest_config.json")
    with open(config_path) as f:
        return json.load(f)


def build_query(bbox, categories):
    south, west, north, east = bbox[1], bbox[0], bbox[3], bbox[2]
    amenity_filter = "|".join(categories)
    return f"""
    [out:json][timeout:300];
    (
      node["amenity"~"{amenity_filter}"]({south},{west},{north},{east});
    );
    out body;
    """


def map_osm_to_poi(element):
    tags = element.tags or {}

    category_map = {
        "cafe": "cafe",
        "restaurant": "restaurant",
        "bar": "bar",
        "pub": "bar",
        "fast_food": "fast_food",
        "food_court": "fast_food",
    }

    category = category_map.get(tags.get("amenity", ""), "cafe")
    subcategory = tags.get("cuisine", None)

    name = tags.get("name", "Unknown")
    name_en = tags.get("name:en", None)
    address_parts = []
    if tags.get("addr:street"):
        address_parts.append(tags["addr:housename"] if tags.get("addr:housename") else "")
        address_parts.append(tags["addr:street"])
        if tags.get("addr:housenumber"):
            address_parts.append(tags["addr:housenumber"])
    address = " ".join(p for p in address_parts if p) or None

    city = tags.get("addr:city", None)
    phone = tags.get("phone", tags.get("contact:phone", None))
    website = tags.get("website", tags.get("contact:website", None))

    opening_hours = tags.get("opening_hours", None)

    rating = 0.0
    review_count = 0
    price_level = 2

    has_wifi = tags.get("internet_access", "") in ("wlan", "yes", "free")
    has_outlets = tags.get("socket", "") == "yes" or tags.get("power_supply", "") == "yes"
    has_terrace = tags.get("outdoor_seating", "") == "yes" or tags.get("terrace", "") == "yes"
    has_parking = tags.get("parking", "") == "yes" or tags.get("amenity:parking", "") == "yes"
    has_live_music = tags.get("music", "") == "live" or tags.get("live_music", "") == "yes"
    has_breakfast = tags.get("breakfast", "") == "yes"
    is_quiet = tags.get("smoking", "") == "no"
    is_family_friendly = tags.get("family_friendly", "") == "yes"
    is_romantic = False
    is_dog_friendly = tags.get("dog", "") == "yes"

    noise_level = "medium"
    if is_quiet:
        noise_level = "low"

    return {
        "osm_id": element.id,
        "osm_type": "node",
        "name": name,
        "name_en": name_en,
        "category": category,
        "subcategory": subcategory,
        "address": address,
        "city": city,
        "phone": phone,
        "website": website,
        "opening_hours": opening_hours,
        "rating": rating,
        "review_count": review_count,
        "price_level": price_level,
        "has_wifi": has_wifi,
        "has_outlets": has_outlets,
        "has_terrace": has_terrace,
        "has_parking": has_parking,
        "has_live_music": has_live_music,
        "has_breakfast": has_breakfast,
        "is_quiet": is_quiet,
        "is_family_friendly": is_family_friendly,
        "is_romantic": is_romantic,
        "is_dog_friendly": is_dog_friendly,
        "noise_level": noise_level,
        "lat": float(element.lat),
        "lon": float(element.lon),
    }


def insert_pois(cursor, pois):
    insert_sql = """
        INSERT INTO pois (
            osm_id, osm_type, name, name_en, category, subcategory,
            address, city, phone, website, opening_hours,
            rating, review_count, price_level,
            has_wifi, has_outlets, has_terrace, has_parking,
            has_live_music, has_breakfast, is_quiet,
            is_family_friendly, is_romantic, is_dog_friendly,
            noise_level, geom
        ) VALUES (
            %(osm_id)s, %(osm_type)s, %(name)s, %(name_en)s, %(category)s, %(subcategory)s,
            %(address)s, %(city)s, %(phone)s, %(website)s, %(opening_hours)s,
            %(rating)s, %(review_count)s, %(price_level)s,
            %(has_wifi)s, %(has_outlets)s, %(has_terrace)s, %(has_parking)s,
            %(has_live_music)s, %(has_breakfast)s, %(is_quiet)s,
            %(is_family_friendly)s, %(is_romantic)s, %(is_dog_friendly)s,
            %(noise_level)s,
            ST_SetSRID(ST_MakePoint(%(lon)s, %(lat)s), 4326)::geography
        )
        ON CONFLICT (osm_id) DO UPDATE SET
            name = EXCLUDED.name,
            name_en = EXCLUDED.name_en,
            category = EXCLUDED.category,
            subcategory = EXCLUDED.subcategory,
            address = EXCLUDED.address,
            city = EXCLUDED.city,
            phone = EXCLUDED.phone,
            website = EXCLUDED.website,
            opening_hours = EXCLUDED.opening_hours,
            has_wifi = EXCLUDED.has_wifi,
            has_outlets = EXCLUDED.has_outlets,
            has_terrace = EXCLUDED.has_terrace,
            has_parking = EXCLUDED.has_parking,
            has_live_music = EXCLUDED.has_live_music,
            has_breakfast = EXCLUDED.has_breakfast,
            is_quiet = EXCLUDED.is_quiet,
            is_family_friendly = EXCLUDED.is_family_friendly,
            is_romantic = EXCLUDED.is_romantic,
            is_dog_friendly = EXCLUDED.is_dog_friendly,
            noise_level = EXCLUDED.noise_level,
            last_updated = NOW()
    """
    cursor.executemany(insert_sql, pois)


def main():
    config = load_config()
    db_url = os.environ.get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable")

    print(f"Fetching POIs from Overpass API for {config['city']}...")
    api = overpy.Overpass()
    query = build_query(config["bbox"], config["categories"])
    result = api.query(query)

    print(f"Found {len(result.nodes)} nodes")

    pois = [map_osm_to_poi(node) for node in result.nodes]

    print(f"Connecting to database...")
    conn = psycopg2.connect(db_url)
    cursor = conn.cursor()

    batch_size = config["batch_size"]
    for i in range(0, len(pois), batch_size):
        batch = pois[i : i + batch_size]
        insert_pois(cursor, batch)
        conn.commit()
        print(f"Inserted batch {i // batch_size + 1} ({len(batch)} POIs)")

    cursor.close()
    conn.close()
    print(f"Done! Total POIs inserted: {len(pois)}")


if __name__ == "__main__":
    main()

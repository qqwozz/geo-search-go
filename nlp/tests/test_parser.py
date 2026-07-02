import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from parser import parse_query


def test_parse_cafe():
    result = parse_query("кафе с террасой")
    assert result["category"] == "cafe"
    assert result["features"].get("terrace") is True


def test_parse_restaurant():
    result = parse_query("ресторан для ужина")
    assert result["category"] == "restaurant"
    assert result["intent"] == "dinner"


def test_parse_work_intent():
    result = parse_query("где поработать с ноутбуком")
    assert result["intent"] == "work"
    assert result["features"].get("wifi") is True


def test_parse_distance():
    result = parse_query("кафе недалеко от метро")
    assert result["radius"] == 1000
    assert result["radius_raw"] == "недалеко"


def test_parse_metro():
    result = parse_query("кафе рядом с метро Парк Культуры")
    assert result["location"] is not None
    assert result["location"]["metro"] == "Парк Культуры"


def test_parse_quiet_wifi():
    result = parse_query("тихое кафе с вайфаем")
    assert result["features"].get("quiet") is True
    assert result["features"].get("wifi") is True


def test_parse_breakfast():
    result = parse_query("где позавтракать")
    assert result["features"].get("breakfast") is True
    assert result["intent"] == "breakfast"


def test_parse_romantic():
    result = parse_query("романтичный ресторан для свидания")
    assert result["intent"] == "romantic"
    assert result["features"].get("romantic") is True


def test_parse_bar():
    result = parse_query("бар с живой музыкой")
    assert result["category"] == "bar"
    assert result["features"].get("live_music") is True


def test_parse_default():
    result = parse_query("что-нибудь вкусное")
    assert result["category"] == "cafe"
    assert result["intent"] == "default"
    assert result["radius"] == 2000

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from location_parser import extract_metro, extract_street


def test_extract_metro_simple():
    assert extract_metro("кафе рядом с метро Парк Культуры") == "Парк Культуры"


def test_extract_metro_with_punctuation():
    assert extract_metro("кафе, метро Тверская, рядом") == "Тверская"


def test_extract_metro_not_found():
    assert extract_metro("кафе с террасой") is None


def test_extract_metro_at_end():
    assert extract_metro("тихое кафе метро Борисово") == "Борисово"


def test_extract_street():
    result = extract_street("кафе на Тверской улице")
    assert result is not None
    assert "Тверской" in result

import re
from typing import Optional

METRO_PATTERN = re.compile(
    r"метро\s+([А-Яа-яёЁ][А-Яа-яёЁ\s]*?)(?:\s*[,.!]|\s*$)"
)

STREET_PATTERN = re.compile(
    r"(?:на|у|около|рядом с)\s+([А-Яа-яёЁ][а-яёЁ]+(?:\s+(?:улица|переулок|проспект|бульвар|шоссе|набережная)))",
    re.IGNORECASE,
)


def extract_metro(text: str) -> Optional[str]:
    match = METRO_PATTERN.search(text)
    if match:
        metro = match.group(1).strip()
        if metro:
            return metro
    return None


def extract_street(text: str) -> Optional[str]:
    match = STREET_PATTERN.search(text)
    if match:
        return match.group(1).strip()
    return None

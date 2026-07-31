"""EvalForge scorers: turn (output, expected) into a 0..1 score."""

from __future__ import annotations

import re
from difflib import SequenceMatcher


def exact_match(output: str, expected: str, **_: object) -> float:
    return 1.0 if output.strip() == expected.strip() else 0.0


def contains(output: str, expected: str, **_: object) -> float:
    return 1.0 if expected.strip().lower() in output.lower() else 0.0


def regex_match(output: str, pattern: str, **_: object) -> float:
    return 1.0 if re.search(pattern, output) else 0.0


def semantic_similarity(output: str, expected: str, **_: object) -> float:
    """Lightweight, dependency-free proxy for semantic closeness.

    Uses token overlap + sequence ratio. Swap for embedding cosine or an
    LLM-as-judge in production; the scorer interface stays identical.
    """
    a, b = output.lower().split(), expected.lower().split()
    if not a or not b:
        return 0.0
    overlap = len(set(a) & set(b)) / len(set(a) | set(b))
    ratio = SequenceMatcher(None, output.lower(), expected.lower()).ratio()
    return round(0.5 * overlap + 0.5 * ratio, 4)


SCORERS = {
    "exact_match": exact_match,
    "contains": contains,
    "regex": regex_match,
    "semantic": semantic_similarity,
}


def get_scorer(name: str):
    if name not in SCORERS:
        raise ValueError(f"unknown scorer '{name}'. available: {list(SCORERS)}")
    return SCORERS[name]

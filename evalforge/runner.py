"""EvalForge runner: load an eval set, score cases, aggregate, gate on threshold.

This is the 'unit tests for prompts' engine. In CI, a failing threshold returns a
non-zero exit code, blocking the deploy -- eval-driven development as a real gate.
"""

from __future__ import annotations

import json
from collections.abc import Callable
from dataclasses import asdict, dataclass
from pathlib import Path

import yaml

from evalforge.scorers import get_scorer


@dataclass
class CaseResult:
    id: str
    scorer: str
    weight: float
    score: float
    passed: bool


@dataclass
class EvalReport:
    aggregate: float
    threshold: float
    passed: bool
    cases: list[CaseResult]

    def to_json(self) -> str:
        return json.dumps(
            {**asdict(self), "cases": [asdict(c) for c in self.cases]}, indent=2
        )


def run_eval(
    eval_set_path: str | Path,
    target: Callable[[str], str],
    threshold: float | None = None,
) -> EvalReport:
    spec = yaml.safe_load(Path(eval_set_path).read_text(encoding="utf-8"))
    threshold = threshold if threshold is not None else float(spec.get("threshold", 0.8))

    results: list[CaseResult] = []
    for case in spec["cases"]:
        output = target(case["input"])
        scorer = get_scorer(case["scorer"])
        kwargs = {k: case[k] for k in ("expected", "pattern") if k in case}
        score = float(scorer(output, **kwargs))
        weight = float(case.get("weight", 1.0))
        case_threshold = float(case.get("pass_score", 0.5))
        results.append(
            CaseResult(
                id=case["id"],
                scorer=case["scorer"],
                weight=weight,
                score=score,
                passed=score >= case_threshold,
            )
        )

    total_w = sum(r.weight for r in results) or 1.0
    aggregate = round(sum(r.score * r.weight for r in results) / total_w, 4)
    return EvalReport(
        aggregate=aggregate,
        threshold=threshold,
        passed=aggregate >= threshold,
        cases=results,
    )

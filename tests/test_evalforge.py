"""EvalForge tests: scorers, aggregation, and the regression-catch gate."""

from pathlib import Path

from evalforge.runner import run_eval
from evalforge.scorers import contains, exact_match, regex_match, semantic_similarity
from examples.demo_target import target

EVAL_SET = Path(__file__).parent.parent / "examples" / "eval_set.yaml"


def test_scorers_basic():
    assert exact_match("abc", "abc") == 1.0
    assert contains("the answer is 30 days", "30 days") == 1.0
    assert regex_match("No discount", "(?i)no") == 1.0
    assert 0.0 <= semantic_similarity("cancel anytime", "you can cancel anytime") <= 1.0


def test_good_target_passes_gate():
    report = run_eval(EVAL_SET, target)
    assert report.passed is True
    assert report.aggregate >= report.threshold


def test_regressed_target_fails_gate():
    """A regressed agent (drops the refund detail) must fail the eval gate."""

    def regressed(question: str) -> str:
        if "cancel" in question.lower():
            return "Yes, cancel anytime."
        return "I don't know."  # lost refund + discount handling

    report = run_eval(EVAL_SET, regressed)
    assert report.passed is False


def test_weights_affect_aggregate():
    report = run_eval(EVAL_SET, target)
    assert any(c.weight == 2.0 for c in report.cases)
    assert 0.0 <= report.aggregate <= 1.0

"""EvalForge CLI: `evalforge run <eval_set.yaml>` -> exits non-zero if below threshold."""

from __future__ import annotations

import argparse
import importlib
import sys

from evalforge.runner import run_eval


def _load_target(spec: str):
    """Load a callable from 'module:function' (defaults to examples.demo_target:target)."""
    module_name, _, func_name = spec.partition(":")
    module = importlib.import_module(module_name)
    return getattr(module, func_name or "target")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="evalforge")
    parser.add_argument("eval_set")
    parser.add_argument("--target", default="examples.demo_target:target")
    parser.add_argument("--threshold", type=float, default=None)
    args = parser.parse_args(argv)

    report = run_eval(args.eval_set, _load_target(args.target), args.threshold)
    print(report.to_json())
    status = "PASS" if report.passed else "FAIL"
    print(f"\n[{status}] aggregate={report.aggregate} threshold={report.threshold}")
    return 0 if report.passed else 1


if __name__ == "__main__":
    sys.exit(main())

# EvalForge

![Python](https://img.shields.io/badge/Python-3.11-3776AB?logo=python&logoColor=white)
![YAML](https://img.shields.io/badge/config-YAML-CB171E?logo=yaml&logoColor=white)
![CI Gate](https://img.shields.io/badge/CI-eval%20gate-blue?logo=githubactions&logoColor=white)
![tests](https://img.shields.io/badge/tests-4%20passing-brightgreen)
![license](https://img.shields.io/badge/license-MIT-green)

**Unit tests for your prompts. A CI gate for LLM quality.**

Teams ship LLM/prompt changes and silently regress quality. EvalForge defines graded eval sets in YAML, scores your agent on every commit, and **fails the build** if quality drops below a threshold - turning eval-driven development into a real CI gate.

## Why it matters

LLM apps have no "unit tests." A prompt tweak that helps one case can quietly break ten others. EvalForge makes quality a **first-class, versioned, enforced** artifact.

## Quickstart

```bash
pip install -r requirements.txt
python -m evalforge.cli examples/eval_set.yaml --target examples.demo_target:target
```

Exit code is `0` if the aggregate score meets the threshold, non-zero otherwise - drop it straight into CI.

## Define an eval set (YAML)

```yaml
threshold: 0.75
cases:
  - id: refund-window
    input: "How many days do I have to request a refund?"
    scorer: contains
    expected: "30 days"
    weight: 2.0
    pass_score: 1.0
```

**Scorers:** `exact_match`, `contains`, `regex`, `semantic` (extensible - add LLM-as-judge, embedding cosine, etc.).

## Point it at your agent

Any `module:function` that maps `str -> str`:

```bash
python -m evalforge.cli my_evals.yaml --target myapp.agent:answer --threshold 0.85
```

## CI: catch regressions before merge

The included GitHub Action runs the eval gate on every PR. See `tests/test_evalforge.py::test_regressed_target_fails_gate` for a demonstration of a regression being caught.

## How it works

```mermaid
flowchart LR
  classDef proc fill:#4a90e2,stroke:#2c5aa0,color:#fff
  classDef good fill:#27ae60,stroke:#1e8449,color:#fff
  classDef bad fill:#e74c3c,stroke:#c0392b,color:#fff
  classDef work fill:#8e44ad,stroke:#6c3483,color:#fff
  ES["eval set<br/>inputs + per-case scorer + pass bar"]:::proc
  RUN["run against the target<br/>(prompt + model)"]:::work
  SCORE["score each case<br/>exact / regex (deterministic)<br/>LLM-as-judge (opt-in)"]:::work
  AGG{"weighted aggregate<br/>vs per-case bars"}:::work
  PASS["pass = exit 0<br/>(deploy proceeds)"]:::good
  FAIL["fail = non-zero exit<br/>(deploy blocked)"]:::bad
  ES --> RUN --> SCORE --> AGG
  AGG -->|meets bar| PASS
  AGG -->|below bar| FAIL
```

## Layout

```
eval-forge/
├── evalforge/
│   ├── runner.py       # loads a YAML eval set, runs each case, aggregates the score
│   ├── cli.py          # `python -m evalforge.cli` — the CI entry point (exit code = the gate)
│   └── scorers/        # exact_match, contains, regex, semantic — add your own here
├── examples/
│   ├── eval_set.yaml   # a worked example eval set
│   └── demo_target.py  # a str -> str target to score against
├── tests/              # incl. a regression that proves the gate fails on a worse target
└── DESIGN.md           # deterministic-first scoring, the two-threshold model, the non-goals
```

## Design notes

- **[DESIGN.md](DESIGN.md)** - why scoring is deterministic-first, the two-threshold
  (per-case + weighted aggregate) design, why the gate is just an exit code, and the
  non-goals (it's a gate, not a scoring model or a dataset tool).

## Part of [parag-labs](https://github.com/parag-labs)

Small, focused tools for building AI systems you can trust.

LedgerRAG · **EvalForge** · AgentGuard · PromptShield · DeployKit

## License

MIT

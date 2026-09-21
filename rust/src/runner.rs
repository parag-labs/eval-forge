//! EvalForge runner: score an eval set against a target and gate on a threshold.

use crate::scorers::{get_scorer, round4};
use serde::Deserialize;
use std::fs;

/// The scored outcome of a single eval case.
#[derive(Debug, Clone, PartialEq)]
pub struct CaseResult {
    /// The case identifier from the eval set.
    pub id: String,
    /// The name of the scorer that graded the case.
    pub scorer: String,
    /// The case's contribution to the weighted aggregate.
    pub weight: f64,
    /// The scorer's 0..1 result for the target output.
    pub score: f64,
    /// Whether `score` met the case's `pass_score` threshold.
    pub passed: bool,
}

/// The aggregate outcome of running an eval set.
#[derive(Debug, Clone, PartialEq)]
pub struct EvalReport {
    /// The weighted mean case score, rounded to four decimals.
    pub aggregate: f64,
    /// The gate the aggregate had to meet to pass.
    pub threshold: f64,
    /// Whether `aggregate` met `threshold`.
    pub passed: bool,
    /// The per-case results in eval-set order.
    pub cases: Vec<CaseResult>,
}

#[derive(Debug, Deserialize)]
struct CaseSpec {
    id: String,
    input: String,
    scorer: String,
    #[serde(default)]
    expected: String,
    #[serde(default)]
    pattern: String,
    weight: Option<f64>,
    pass_score: Option<f64>,
}

#[derive(Debug, Deserialize)]
struct EvalSpec {
    threshold: Option<f64>,
    #[serde(default)]
    cases: Vec<CaseSpec>,
}

/// Loads the eval set at `eval_set_path`, scores each case against `target`, and
/// gates the weighted-mean aggregate on a threshold. A `Some` threshold
/// overrides the eval set's own threshold (which defaults to 0.8 when absent).
/// This is the "unit tests for prompts" engine: in CI a failing gate is meant to
/// block the deploy.
pub fn run_eval(
    eval_set_path: &str,
    target: impl Fn(&str) -> String,
    threshold: Option<f64>,
) -> Result<EvalReport, String> {
    let data = fs::read_to_string(eval_set_path).map_err(|e| e.to_string())?;
    let spec: EvalSpec = serde_yaml::from_str(&data).map_err(|e| e.to_string())?;

    let mut effective = spec.threshold.unwrap_or(0.8);
    if let Some(t) = threshold {
        effective = t;
    }

    let mut results: Vec<CaseResult> = Vec::with_capacity(spec.cases.len());
    for c in &spec.cases {
        let output = target(&c.input);
        let scorer = get_scorer(&c.scorer)?;
        let arg = if c.scorer == "regex" {
            &c.pattern
        } else {
            &c.expected
        };
        let score = scorer(&output, arg);
        let weight = c.weight.unwrap_or(1.0);
        let case_threshold = c.pass_score.unwrap_or(0.5);
        results.push(CaseResult {
            id: c.id.clone(),
            scorer: c.scorer.clone(),
            weight,
            score,
            passed: score >= case_threshold,
        });
    }

    let mut total_w: f64 = results.iter().map(|r| r.weight).sum();
    if total_w == 0.0 {
        total_w = 1.0;
    }
    let weighted: f64 = results.iter().map(|r| r.score * r.weight).sum();
    let aggregate = round4(weighted / total_w);

    Ok(EvalReport {
        aggregate,
        threshold: effective,
        passed: aggregate >= effective,
        cases: results,
    })
}

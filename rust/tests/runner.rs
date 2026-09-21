use eval_forge::{demo_target, run_eval};
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};

fn eval_set_path() -> String {
    let mut dir = std::env::current_dir().expect("cwd");
    loop {
        let candidate = dir.join("examples").join("eval_set.yaml");
        if candidate.exists() {
            return candidate.to_string_lossy().into_owned();
        }
        if !dir.pop() {
            panic!("examples/eval_set.yaml not found walking up from working dir");
        }
    }
}

static COUNTER: AtomicU64 = AtomicU64::new(0);

/// Writes an eval set into the crate directory (never a system temp dir) and
/// removes it on drop.
struct TempEval {
    path: PathBuf,
}

impl TempEval {
    fn new(content: &str) -> TempEval {
        let n = COUNTER.fetch_add(1, Ordering::SeqCst);
        let name = format!("eval-test-{}-{}.yaml", std::process::id(), n);
        let path = std::env::current_dir().expect("cwd").join(name);
        std::fs::write(&path, content).expect("write temp eval");
        TempEval { path }
    }

    fn path(&self) -> String {
        self.path.to_string_lossy().into_owned()
    }
}

impl Drop for TempEval {
    fn drop(&mut self) {
        let _ = std::fs::remove_file(&self.path);
    }
}

fn approx(a: f64, b: f64) {
    assert!((a - b).abs() < 1e-9, "expected {b}, got {a}");
}

#[test]
fn good_target_passes_gate() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    assert!(report.passed, "good target should pass the gate");
    assert!(report.aggregate >= report.threshold);
}

#[test]
fn aggregate_has_exact_reference_value() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    approx(report.aggregate, 0.9238);
    approx(report.threshold, 0.75);
}

#[test]
fn three_cases_in_eval_set_order() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    let ids: Vec<&str> = report.cases.iter().map(|c| c.id.as_str()).collect();
    assert_eq!(
        ids,
        vec!["refund-window", "cancellation", "no-hallucinated-discount"]
    );
}

#[test]
fn case_weights_and_scorers_match_spec() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    let weights: Vec<f64> = report.cases.iter().map(|c| c.weight).collect();
    let scorers: Vec<&str> = report.cases.iter().map(|c| c.scorer.as_str()).collect();
    assert_eq!(weights, vec![2.0, 1.0, 1.5]);
    assert_eq!(scorers, vec!["contains", "semantic", "regex"]);
}

#[test]
fn weights_affect_aggregate() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    assert!(report.cases.iter().any(|c| c.weight == 2.0));
    assert!((0.0..=1.0).contains(&report.aggregate));
}

#[test]
fn per_case_scores_and_pass_flags() {
    let report = run_eval(&eval_set_path(), demo_target, None).unwrap();
    approx(report.cases[0].score, 1.0);
    assert!(report.cases[0].passed);
    approx(report.cases[1].score, 0.6573);
    assert!(report.cases[1].passed);
    approx(report.cases[2].score, 1.0);
    assert!(report.cases[2].passed);
}

#[test]
fn regressed_target_fails_gate() {
    let regressed = |q: &str| {
        if q.to_lowercase().contains("cancel") {
            "Yes, cancel anytime.".to_string()
        } else {
            "I don't know.".to_string()
        }
    };
    let report = run_eval(&eval_set_path(), regressed, None).unwrap();
    assert!(!report.passed, "regressed target should fail the gate");
}

#[test]
fn threshold_override_higher_fails() {
    let report = run_eval(&eval_set_path(), demo_target, Some(0.99)).unwrap();
    assert!(!report.passed);
    approx(report.threshold, 0.99);
}

#[test]
fn threshold_override_lower_passes() {
    let report = run_eval(&eval_set_path(), demo_target, Some(0.1)).unwrap();
    assert!(report.passed);
    approx(report.threshold, 0.1);
}

#[test]
fn missing_file_is_err() {
    assert!(run_eval("does-not-exist.yaml", demo_target, None).is_err());
}

#[test]
fn unknown_scorer_is_err() {
    let tmp = TempEval::new(
        "threshold: 0.5\ncases:\n  - id: c1\n    input: hello\n    scorer: bogus\n    expected: hello\n",
    );
    assert!(run_eval(&tmp.path(), demo_target, None).is_err());
}

#[test]
fn zero_weights_fall_back_to_one() {
    let tmp = TempEval::new(
        "threshold: 0.0\ncases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n    weight: 0.0\n  - id: b\n    input: y\n    scorer: exact_match\n    expected: y\n    weight: 0.0\n",
    );
    let report = run_eval(&tmp.path(), demo_target, None).unwrap();
    approx(report.aggregate, 0.0);
    assert!(report.passed);
}

#[test]
fn default_threshold_when_absent() {
    let tmp = TempEval::new(
        "cases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n",
    );
    let echo = |s: &str| s.to_string();
    let report = run_eval(&tmp.path(), echo, None).unwrap();
    approx(report.threshold, 0.8);
    approx(report.aggregate, 1.0);
}

#[test]
fn demo_target_branches() {
    assert_eq!(
        demo_target("How many days for a refund?"),
        "You have 30 days from purchase to request a full refund."
    );
    assert_eq!(
        demo_target("Can I cancel?"),
        "Yes, you can cancel your subscription at any time; it ends at the cycle close."
    );
    assert_eq!(
        demo_target("Any discount?"),
        "No, we do not offer a 90% loyalty discount."
    );
    assert_eq!(
        demo_target("What is the weather?"),
        "I don't have information on that."
    );
}

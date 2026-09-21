//! EvalForge: a tiny prompt-evaluation harness — scorers plus a threshold gate.
//!
//! The crate mirrors the Python reference implementation: the same scorers, the
//! same weighted-mean aggregate, and the same pass/fail gate so an eval set is
//! "unit tests for prompts" that can block a deploy in CI.

pub mod demo_target;
pub mod runner;
pub mod scorers;

pub use demo_target::demo_target;
pub use runner::{run_eval, CaseResult, EvalReport};
pub use scorers::{
    contains, exact_match, get_scorer, is_known_scorer, regex_match, round4, scorer_names,
    semantic_similarity, Scorer, SCORER_NAMES,
};

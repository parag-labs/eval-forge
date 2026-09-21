//! A tiny deterministic demo agent used by the example eval set.

/// Answers a support question with a canned response; replace it with a real LLM
/// app callable. One branch is deliberately wrong-but-plausible so an eval set
/// has something to catch.
pub fn demo_target(question: &str) -> String {
    let q = question.to_lowercase();
    if q.contains("refund") {
        "You have 30 days from purchase to request a full refund.".to_string()
    } else if q.contains("cancel") {
        "Yes, you can cancel your subscription at any time; it ends at the cycle close.".to_string()
    } else if q.contains("discount") {
        "No, we do not offer a 90% loyalty discount.".to_string()
    } else {
        "I don't have information on that.".to_string()
    }
}

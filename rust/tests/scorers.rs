use eval_forge::{
    contains, exact_match, get_scorer, is_known_scorer, regex_match, round4, scorer_names,
    semantic_similarity, SCORER_NAMES,
};

fn approx(a: f64, b: f64) {
    assert!((a - b).abs() < 1e-9, "expected {b}, got {a}");
}

#[test]
fn exact_match_identical() {
    approx(exact_match("hello", "hello"), 1.0);
}

#[test]
fn exact_match_trims_surrounding_whitespace() {
    approx(exact_match("  hello  ", "hello"), 1.0);
    approx(exact_match("hello", "\thello\n"), 1.0);
}

#[test]
fn exact_match_different() {
    approx(exact_match("hello", "world"), 0.0);
}

#[test]
fn exact_match_is_case_sensitive() {
    approx(exact_match("Hello", "hello"), 0.0);
}

#[test]
fn contains_substring_present() {
    approx(contains("the quick brown fox", "quick"), 1.0);
}

#[test]
fn contains_is_case_insensitive() {
    approx(contains("The Quick Brown FOX", "quick brown"), 1.0);
}

#[test]
fn contains_trims_expected() {
    approx(contains("refund in 30 days", "  30 days  "), 1.0);
}

#[test]
fn contains_absent() {
    approx(contains("the quick brown fox", "lazy dog"), 0.0);
}

#[test]
fn contains_empty_expected_matches() {
    approx(contains("anything", ""), 1.0);
}

#[test]
fn regex_matches_anywhere() {
    approx(regex_match("abc123def", r"\d+"), 1.0);
}

#[test]
fn regex_inline_case_insensitive_flag() {
    approx(regex_match("NO WAY", "(?i)no"), 1.0);
    approx(
        regex_match("Definitely not", "(?i)(no|not|don't|do not)"),
        1.0,
    );
}

#[test]
fn regex_no_match() {
    approx(regex_match("abcdef", r"\d+"), 0.0);
}

#[test]
fn regex_anchors_respected() {
    approx(regex_match("xabc", "^abc"), 0.0);
    approx(regex_match("abc", "^abc$"), 1.0);
}

#[test]
fn semantic_identical_strings() {
    approx(semantic_similarity("abc def", "abc def"), 1.0);
}

#[test]
fn semantic_empty_output_is_zero() {
    approx(semantic_similarity("", "x"), 0.0);
}

#[test]
fn semantic_empty_expected_is_zero() {
    approx(semantic_similarity("x", ""), 0.0);
}

#[test]
fn semantic_both_empty_is_zero() {
    approx(semantic_similarity("", ""), 0.0);
}

#[test]
fn semantic_reordered_tokens() {
    // set overlap is 1.0; sequence ratio differs because order changed.
    approx(semantic_similarity("hello world", "world hello"), 0.7273);
}

#[test]
fn semantic_partial_overlap() {
    approx(
        semantic_similarity("the quick brown fox", "a quick brown dog"),
        0.5556,
    );
}

#[test]
fn semantic_repeated_chars() {
    approx(semantic_similarity("aaaa", "aa"), 0.3333);
}

#[test]
fn semantic_reference_cancellation_value() {
    let out = "Yes, you can cancel your subscription at any time; it ends at the cycle close.";
    let exp = "Yes, you can cancel your subscription at any time.";
    approx(semantic_similarity(out, exp), 0.6573);
}

#[test]
fn semantic_reference_short_value() {
    approx(
        semantic_similarity("cancel anytime", "you can cancel anytime"),
        0.6389,
    );
}

#[test]
fn semantic_stays_within_unit_interval() {
    let v = semantic_similarity("the quick brown fox jumps", "the lazy dog sleeps soundly");
    assert!((0.0..=1.0).contains(&v), "score {v} out of range");
}

#[test]
fn round4_half_to_even() {
    approx(round4(0.123449), 0.1234);
    approx(round4(1.0 / 3.0), 0.3333);
}

#[test]
fn get_scorer_returns_each_builtin() {
    for name in SCORER_NAMES {
        assert!(get_scorer(name).is_ok(), "{name} should resolve");
    }
}

#[test]
fn get_scorer_unknown_is_err() {
    let err = get_scorer("nope").unwrap_err();
    assert!(err.contains("unknown scorer 'nope'"), "message was {err}");
}

#[test]
fn is_known_scorer_reports_membership() {
    assert!(is_known_scorer("semantic"));
    assert!(is_known_scorer("exact_match"));
    assert!(!is_known_scorer("fuzzy"));
}

#[test]
fn scorer_names_lists_all_four_in_order() {
    assert_eq!(
        scorer_names(),
        vec!["exact_match", "contains", "regex", "semantic"]
    );
}

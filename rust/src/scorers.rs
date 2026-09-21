//! EvalForge scorers: turn a target output and an argument into a 0..1 score.

use regex::Regex;
use std::collections::HashMap;
use std::collections::HashSet;

/// A scorer turns a target output and a single argument (the expected string,
/// or the pattern for the regex scorer) into a score in the range 0..1.
pub type Scorer = fn(&str, &str) -> f64;

/// Scores 1.0 when output and expected are equal after trimming surrounding
/// whitespace, else 0.0.
pub fn exact_match(output: &str, expected: &str) -> f64 {
    if output.trim() == expected.trim() {
        1.0
    } else {
        0.0
    }
}

/// Scores 1.0 when expected (trimmed and lower-cased) appears anywhere in
/// output (lower-cased), else 0.0.
pub fn contains(output: &str, expected: &str) -> f64 {
    let needle = expected.trim().to_lowercase();
    if output.to_lowercase().contains(needle.as_str()) {
        1.0
    } else {
        0.0
    }
}

/// Scores 1.0 when pattern matches anywhere in output, else 0.0. The pattern
/// supports inline flags such as `(?i)`; an invalid pattern panics, mirroring
/// the Python reference raising `re.error`.
pub fn regex_match(output: &str, pattern: &str) -> f64 {
    let re = Regex::new(pattern).expect("invalid regex pattern");
    if re.is_match(output) {
        1.0
    } else {
        0.0
    }
}

/// A lightweight, dependency-free proxy for semantic closeness: the mean of
/// Jaccard token overlap and difflib's sequence ratio, rounded to four
/// decimals. Not a real embedding model; the scorer interface is identical so
/// it can be swapped for one in production.
pub fn semantic_similarity(output: &str, expected: &str) -> f64 {
    let lo = output.to_lowercase();
    let le = expected.to_lowercase();
    let a: Vec<&str> = lo.split_whitespace().collect();
    let b: Vec<&str> = le.split_whitespace().collect();
    if a.is_empty() || b.is_empty() {
        return 0.0;
    }
    let sa: HashSet<&str> = a.iter().copied().collect();
    let sb: HashSet<&str> = b.iter().copied().collect();
    let inter = sa.intersection(&sb).count();
    let union = sa.union(&sb).count();
    let overlap = inter as f64 / union as f64;
    let ratio = sequence_ratio(&lo, &le);
    round4(0.5 * overlap + 0.5 * ratio)
}

/// Rounds v to four decimal places using round-half-to-even, reproducing the
/// value Python's `round(v, 4)` yields for the same input.
pub fn round4(v: f64) -> f64 {
    format!("{v:.4}").parse::<f64>().unwrap()
}

/// Reproduces `difflib.SequenceMatcher(None, a, b).ratio()`: it is `2*M/T` where
/// `T` is the combined length of both strings and `M` is the total size of the
/// matching blocks found by the Ratcliff/Obershelp algorithm. Comparison is per
/// Unicode code point, and the autojunk heuristic is not applied (it only
/// affects sequences of 200+ elements, well beyond the short strings scored
/// here).
fn sequence_ratio(a: &str, b: &str) -> f64 {
    let ra: Vec<char> = a.chars().collect();
    let rb: Vec<char> = b.chars().collect();
    if ra.is_empty() && rb.is_empty() {
        return 1.0;
    }
    let mut b2j: HashMap<char, Vec<usize>> = HashMap::new();
    for (j, &ch) in rb.iter().enumerate() {
        b2j.entry(ch).or_default().push(j);
    }
    let matches = total_matches(&ra, &b2j, 0, ra.len(), 0, rb.len());
    2.0 * matches as f64 / (ra.len() + rb.len()) as f64
}

fn total_matches(
    a: &[char],
    b2j: &HashMap<char, Vec<usize>>,
    alo: usize,
    ahi: usize,
    blo: usize,
    bhi: usize,
) -> usize {
    let (i, j, k) = find_longest_match(a, b2j, alo, ahi, blo, bhi);
    if k == 0 {
        return 0;
    }
    k + total_matches(a, b2j, alo, i, blo, j) + total_matches(a, b2j, i + k, ahi, j + k, bhi)
}

fn find_longest_match(
    a: &[char],
    b2j: &HashMap<char, Vec<usize>>,
    alo: usize,
    ahi: usize,
    blo: usize,
    bhi: usize,
) -> (usize, usize, usize) {
    let mut besti = alo;
    let mut bestj = blo;
    let mut bestsize = 0usize;
    let mut j2len: HashMap<usize, usize> = HashMap::new();
    for (i, item) in a.iter().enumerate().take(ahi).skip(alo) {
        let mut newj2len: HashMap<usize, usize> = HashMap::new();
        if let Some(js) = b2j.get(item) {
            for &j in js {
                if j < blo {
                    continue;
                }
                if j >= bhi {
                    break;
                }
                let prev = if j == 0 {
                    0
                } else {
                    *j2len.get(&(j - 1)).unwrap_or(&0)
                };
                let k = prev + 1;
                newj2len.insert(j, k);
                if k > bestsize {
                    besti = i + 1 - k;
                    bestj = j + 1 - k;
                    bestsize = k;
                }
            }
        }
        j2len = newj2len;
    }
    (besti, bestj, bestsize)
}

/// The built-in scorer names, in registration order.
pub const SCORER_NAMES: [&str; 4] = ["exact_match", "contains", "regex", "semantic"];

/// Returns the names of the built-in scorers.
pub fn scorer_names() -> Vec<String> {
    SCORER_NAMES.iter().map(|s| s.to_string()).collect()
}

/// Reports whether name is a built-in scorer.
pub fn is_known_scorer(name: &str) -> bool {
    SCORER_NAMES.contains(&name)
}

/// Returns the scorer registered under name, or an error if no such scorer
/// exists, mirroring the Python reference's `get_scorer`.
pub fn get_scorer(name: &str) -> Result<Scorer, String> {
    match name {
        "exact_match" => Ok(exact_match),
        "contains" => Ok(contains),
        "regex" => Ok(regex_match),
        "semantic" => Ok(semantic_similarity),
        _ => Err(format!(
            "unknown scorer '{name}'. available: {SCORER_NAMES:?}"
        )),
    }
}

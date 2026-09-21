import { describe, expect, it } from "vitest";
import {
  contains,
  exactMatch,
  getScorer,
  isKnownScorer,
  regexMatch,
  round4,
  SCORER_NAMES,
  scorerNames,
  semanticSimilarity,
} from "./scorers";

describe("exactMatch", () => {
  it("scores identical strings 1.0", () => {
    expect(exactMatch("hello", "hello")).toBe(1.0);
  });
  it("trims surrounding whitespace on both sides", () => {
    expect(exactMatch("  hello  ", "hello")).toBe(1.0);
    expect(exactMatch("hello", "\thello\n")).toBe(1.0);
  });
  it("scores different strings 0.0", () => {
    expect(exactMatch("hello", "world")).toBe(0.0);
  });
  it("is case sensitive", () => {
    expect(exactMatch("Hello", "hello")).toBe(0.0);
  });
});

describe("contains", () => {
  it("finds a substring", () => {
    expect(contains("the quick brown fox", "quick")).toBe(1.0);
  });
  it("is case insensitive", () => {
    expect(contains("The Quick Brown FOX", "quick brown")).toBe(1.0);
  });
  it("trims the expected needle", () => {
    expect(contains("refund in 30 days", "  30 days  ")).toBe(1.0);
  });
  it("scores 0.0 when absent", () => {
    expect(contains("the quick brown fox", "lazy dog")).toBe(0.0);
  });
  it("matches an empty needle", () => {
    expect(contains("anything", "")).toBe(1.0);
  });
});

describe("regexMatch", () => {
  it("matches anywhere in the output", () => {
    expect(regexMatch("abc123def", "\\d+")).toBe(1.0);
  });
  it("honours a leading (?i) inline flag", () => {
    expect(regexMatch("NO WAY", "(?i)no")).toBe(1.0);
    expect(regexMatch("Definitely not", "(?i)(no|not|don't|do not)")).toBe(1.0);
  });
  it("scores 0.0 on no match", () => {
    expect(regexMatch("abcdef", "\\d+")).toBe(0.0);
  });
  it("respects anchors", () => {
    expect(regexMatch("xabc", "^abc")).toBe(0.0);
    expect(regexMatch("abc", "^abc$")).toBe(1.0);
  });
});

describe("semanticSimilarity", () => {
  it("scores identical strings 1.0", () => {
    expect(semanticSimilarity("abc def", "abc def")).toBe(1.0);
  });
  it("scores empty output 0.0", () => {
    expect(semanticSimilarity("", "x")).toBe(0.0);
  });
  it("scores empty expected 0.0", () => {
    expect(semanticSimilarity("x", "")).toBe(0.0);
  });
  it("scores both empty 0.0", () => {
    expect(semanticSimilarity("", "")).toBe(0.0);
  });
  it("reproduces the reordered-token reference value", () => {
    expect(semanticSimilarity("hello world", "world hello")).toBeCloseTo(0.7273, 10);
  });
  it("reproduces the partial-overlap reference value", () => {
    expect(semanticSimilarity("the quick brown fox", "a quick brown dog")).toBeCloseTo(0.5556, 10);
  });
  it("reproduces the repeated-char reference value", () => {
    expect(semanticSimilarity("aaaa", "aa")).toBeCloseTo(0.3333, 10);
  });
  it("reproduces the cancellation reference value", () => {
    const out = "Yes, you can cancel your subscription at any time; it ends at the cycle close.";
    const exp = "Yes, you can cancel your subscription at any time.";
    expect(semanticSimilarity(out, exp)).toBeCloseTo(0.6573, 10);
  });
  it("reproduces the short reference value", () => {
    expect(semanticSimilarity("cancel anytime", "you can cancel anytime")).toBeCloseTo(0.6389, 10);
  });
  it("stays within [0,1]", () => {
    const v = semanticSimilarity("the quick brown fox jumps", "the lazy dog sleeps soundly");
    expect(v).toBeGreaterThanOrEqual(0.0);
    expect(v).toBeLessThanOrEqual(1.0);
  });
});

describe("round4", () => {
  it("rounds to four decimal places", () => {
    expect(round4(1 / 3)).toBe(0.3333);
    expect(round4(0.123449)).toBe(0.1234);
  });
});

describe("scorer registry", () => {
  it("resolves every built-in scorer", () => {
    for (const name of SCORER_NAMES) {
      expect(typeof getScorer(name)).toBe("function");
    }
  });
  it("throws for an unknown scorer", () => {
    expect(() => getScorer("nope")).toThrow(/unknown scorer 'nope'/);
  });
  it("reports membership", () => {
    expect(isKnownScorer("semantic")).toBe(true);
    expect(isKnownScorer("exact_match")).toBe(true);
    expect(isKnownScorer("fuzzy")).toBe(false);
  });
  it("lists all four names in registration order", () => {
    expect(scorerNames()).toEqual(["exact_match", "contains", "regex", "semantic"]);
  });
});

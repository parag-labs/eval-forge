package evalforge

import (
	"math"
	"testing"
)

func approxEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExactMatchIdentical(t *testing.T) {
	if got := ExactMatch("abc", "abc"); got != 1.0 {
		t.Fatalf("ExactMatch identical = %v, want 1.0", got)
	}
}

func TestExactMatchTrimsSurroundingWhitespace(t *testing.T) {
	if got := ExactMatch("  abc\n", "abc"); got != 1.0 {
		t.Fatalf("ExactMatch with surrounding whitespace = %v, want 1.0", got)
	}
}

func TestExactMatchDifferent(t *testing.T) {
	if got := ExactMatch("abc", "abd"); got != 0.0 {
		t.Fatalf("ExactMatch different = %v, want 0.0", got)
	}
}

func TestExactMatchIsCaseSensitive(t *testing.T) {
	if got := ExactMatch("ABC", "abc"); got != 0.0 {
		t.Fatalf("ExactMatch is case sensitive, got %v, want 0.0", got)
	}
}

func TestContainsSubstring(t *testing.T) {
	if got := Contains("the answer is 30 days", "30 days"); got != 1.0 {
		t.Fatalf("Contains substring = %v, want 1.0", got)
	}
}

func TestContainsIsCaseInsensitive(t *testing.T) {
	if got := Contains("The Answer Is 30 DAYS", "30 days"); got != 1.0 {
		t.Fatalf("Contains should be case-insensitive, got %v, want 1.0", got)
	}
}

func TestContainsTrimsExpected(t *testing.T) {
	if got := Contains("abc def ghi", "  def  "); got != 1.0 {
		t.Fatalf("Contains should trim expected, got %v, want 1.0", got)
	}
}

func TestContainsAbsent(t *testing.T) {
	if got := Contains("hello world", "goodbye"); got != 0.0 {
		t.Fatalf("Contains absent = %v, want 0.0", got)
	}
}

func TestRegexMatches(t *testing.T) {
	if got := RegexMatch("No discount", "(?i)no"); got != 1.0 {
		t.Fatalf("RegexMatch (?i)no = %v, want 1.0", got)
	}
}

func TestRegexInlineCaseInsensitiveAlternation(t *testing.T) {
	if got := RegexMatch("No, we do not offer a 90% loyalty discount.", "(?i)(no|not|don't|do not)"); got != 1.0 {
		t.Fatalf("RegexMatch alternation = %v, want 1.0", got)
	}
}

func TestRegexNoMatch(t *testing.T) {
	if got := RegexMatch("yes absolutely", "(?i)no"); got != 0.0 {
		t.Fatalf("RegexMatch no match = %v, want 0.0", got)
	}
}

func TestRegexSearchesAnywhere(t *testing.T) {
	if got := RegexMatch("the total is 30 days", `\d+ days`); got != 1.0 {
		t.Fatalf("RegexMatch should search anywhere, got %v, want 1.0", got)
	}
}

func TestSemanticIdenticalIsOne(t *testing.T) {
	approxEqual(t, SemanticSimilarity("abc def", "abc def"), 1.0)
}

func TestSemanticEmptyOutputIsZero(t *testing.T) {
	approxEqual(t, SemanticSimilarity("", "something"), 0.0)
}

func TestSemanticEmptyExpectedIsZero(t *testing.T) {
	approxEqual(t, SemanticSimilarity("something", ""), 0.0)
}

func TestSemanticBothEmptyIsZero(t *testing.T) {
	approxEqual(t, SemanticSimilarity("", ""), 0.0)
}

// Reference values captured from the Python difflib-based implementation.
func TestSemanticKnownValues(t *testing.T) {
	cases := []struct {
		output, expected string
		want             float64
	}{
		{"cancel anytime", "you can cancel anytime", 0.6389},
		{"Yes, you can cancel your subscription at any time; it ends at the cycle close.", "Yes, you can cancel your subscription at any time.", 0.6573},
		{"hello world", "world hello", 0.7273},
		{"the quick brown fox", "a quick brown dog", 0.5556},
		{"aaaa", "aa", 0.3333},
	}
	for _, c := range cases {
		approxEqual(t, SemanticSimilarity(c.output, c.expected), c.want)
	}
}

func TestSemanticStaysInUnitRange(t *testing.T) {
	got := SemanticSimilarity("cancel anytime", "you can cancel anytime")
	if got < 0.0 || got > 1.0 {
		t.Fatalf("semantic score %v out of [0,1]", got)
	}
}

func TestGetScorerReturnsEachBuiltin(t *testing.T) {
	for _, name := range []string{"exact_match", "contains", "regex", "semantic"} {
		fn, err := GetScorer(name)
		if err != nil {
			t.Fatalf("GetScorer(%q) unexpected error: %v", name, err)
		}
		if fn == nil {
			t.Fatalf("GetScorer(%q) returned nil func", name)
		}
	}
}

func TestGetScorerUnknownErrors(t *testing.T) {
	_, err := GetScorer("nonsense")
	if err == nil {
		t.Fatalf("GetScorer(nonsense) should error")
	}
}

func TestIsKnownScorer(t *testing.T) {
	if !IsKnownScorer("semantic") {
		t.Fatalf("semantic should be known")
	}
	if IsKnownScorer("bogus") {
		t.Fatalf("bogus should not be known")
	}
}

func TestScorerNamesReturnsCopy(t *testing.T) {
	names := ScorerNames()
	if len(names) != 4 {
		t.Fatalf("expected 4 scorer names, got %d", len(names))
	}
	names[0] = "mutated"
	if ScorerNames()[0] == "mutated" {
		t.Fatalf("ScorerNames must return a defensive copy")
	}
}

package evalforge

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ScorerFunc turns a target output and a single argument (the expected string,
// or the pattern for the regex scorer) into a score in the range 0..1.
type ScorerFunc func(output, arg string) float64

// ExactMatch scores 1.0 when output and expected are equal after trimming
// surrounding whitespace, else 0.0.
func ExactMatch(output, expected string) float64 {
	if strings.TrimSpace(output) == strings.TrimSpace(expected) {
		return 1.0
	}
	return 0.0
}

// Contains scores 1.0 when expected (trimmed and lower-cased) appears anywhere
// in output (lower-cased), else 0.0.
func Contains(output, expected string) float64 {
	if strings.Contains(strings.ToLower(output), strings.ToLower(strings.TrimSpace(expected))) {
		return 1.0
	}
	return 0.0
}

// RegexMatch scores 1.0 when pattern matches anywhere in output, else 0.0. The
// pattern uses RE2 syntax and supports inline flags such as (?i); an invalid
// pattern panics, mirroring the Python reference raising re.error.
func RegexMatch(output, pattern string) float64 {
	if regexp.MustCompile(pattern).MatchString(output) {
		return 1.0
	}
	return 0.0
}

// SemanticSimilarity is a lightweight, dependency-free proxy for semantic
// closeness: the mean of Jaccard token overlap and difflib's sequence ratio,
// rounded to four decimals. It is not a real embedding model; the scorer
// interface is identical so it can be swapped for one in production.
func SemanticSimilarity(output, expected string) float64 {
	a := strings.Fields(strings.ToLower(output))
	b := strings.Fields(strings.ToLower(expected))
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	sa := toSet(a)
	sb := toSet(b)
	inter := 0
	for w := range sa {
		if _, ok := sb[w]; ok {
			inter++
		}
	}
	union := len(sa)
	for w := range sb {
		if _, ok := sa[w]; !ok {
			union++
		}
	}
	overlap := float64(inter) / float64(union)
	ratio := sequenceRatio(strings.ToLower(output), strings.ToLower(expected))
	return round4(0.5*overlap + 0.5*ratio)
}

func toSet(words []string) map[string]struct{} {
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	return set
}

// round4 rounds v to four decimal places using round-half-to-even, reproducing
// the value Python's round(v, 4) yields for the same input.
func round4(v float64) float64 {
	r, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', 4, 64), 64)
	return r
}

// sequenceRatio reproduces difflib.SequenceMatcher(None, a, b).ratio(): it is
// 2*M/T where T is the combined length of both strings and M is the total size
// of the matching blocks found by the Ratcliff/Obershelp algorithm. Comparison
// is per Unicode code point, and the autojunk heuristic is not applied (it only
// affects sequences of 200+ elements, well beyond the short strings scored here).
func sequenceRatio(a, b string) float64 {
	ra := []rune(a)
	rb := []rune(b)
	if len(ra)+len(rb) == 0 {
		return 1.0
	}
	b2j := make(map[rune][]int, len(rb))
	for j, ch := range rb {
		b2j[ch] = append(b2j[ch], j)
	}
	matches := totalMatches(ra, b2j, 0, len(ra), 0, len(rb))
	return 2.0 * float64(matches) / float64(len(ra)+len(rb))
}

func totalMatches(a []rune, b2j map[rune][]int, alo, ahi, blo, bhi int) int {
	i, j, k := findLongestMatch(a, b2j, alo, ahi, blo, bhi)
	if k == 0 {
		return 0
	}
	return k +
		totalMatches(a, b2j, alo, i, blo, j) +
		totalMatches(a, b2j, i+k, ahi, j+k, bhi)
}

func findLongestMatch(a []rune, b2j map[rune][]int, alo, ahi, blo, bhi int) (int, int, int) {
	besti, bestj, bestsize := alo, blo, 0
	j2len := make(map[int]int)
	for i := alo; i < ahi; i++ {
		newj2len := make(map[int]int)
		for _, j := range b2j[a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}
			k := j2len[j-1] + 1
			newj2len[j] = k
			if k > bestsize {
				besti = i - k + 1
				bestj = j - k + 1
				bestsize = k
			}
		}
		j2len = newj2len
	}
	return besti, bestj, bestsize
}

// scorerNames lists the built-in scorers in registration order.
var scorerNames = []string{"exact_match", "contains", "regex", "semantic"}

// ScorerNames returns the names of the built-in scorers.
func ScorerNames() []string {
	out := make([]string, len(scorerNames))
	copy(out, scorerNames)
	return out
}

// IsKnownScorer reports whether name is a built-in scorer.
func IsKnownScorer(name string) bool {
	for _, n := range scorerNames {
		if n == name {
			return true
		}
	}
	return false
}

// GetScorer returns the scorer function registered under name, or an error if no
// such scorer exists, mirroring the Python reference's get_scorer.
func GetScorer(name string) (ScorerFunc, error) {
	switch name {
	case "exact_match":
		return ExactMatch, nil
	case "contains":
		return Contains, nil
	case "regex":
		return RegexMatch, nil
	case "semantic":
		return SemanticSimilarity, nil
	default:
		return nil, fmt.Errorf("unknown scorer %q. available: %v", name, ScorerNames())
	}
}

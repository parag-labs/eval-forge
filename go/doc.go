// Package evalforge is the "unit tests for prompts" engine: it loads a graded
// eval set, scores each case's target output with a named scorer, aggregates the
// weighted scores and gates the result on a threshold.
//
// It is a faithful port of the Python reference package. The scorers reproduce
// Python's difflib.SequenceMatcher.ratio() (Ratcliff/Obershelp) so semantic
// scores are identical across every language port, and aggregate scores are
// rounded to four decimals with round-half-to-even to match Python's round().
package evalforge

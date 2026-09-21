package evalforge

import (
	"os"

	"gopkg.in/yaml.v3"
)

// CaseResult is the scored outcome of a single eval case.
type CaseResult struct {
	// ID is the case identifier from the eval set.
	ID string
	// Scorer is the name of the scorer that graded the case.
	Scorer string
	// Weight is the case's contribution to the weighted aggregate.
	Weight float64
	// Score is the scorer's 0..1 result for the target output.
	Score float64
	// Passed reports whether Score met the case's pass_score threshold.
	Passed bool
}

// EvalReport is the aggregate outcome of running an eval set.
type EvalReport struct {
	// Aggregate is the weighted mean case score, rounded to four decimals.
	Aggregate float64
	// Threshold is the gate the aggregate had to meet to pass.
	Threshold float64
	// Passed reports whether Aggregate met Threshold.
	Passed bool
	// Cases holds the per-case results in eval-set order.
	Cases []CaseResult
}

type caseSpec struct {
	ID        string   `yaml:"id"`
	Input     string   `yaml:"input"`
	Scorer    string   `yaml:"scorer"`
	Expected  string   `yaml:"expected"`
	Pattern   string   `yaml:"pattern"`
	Weight    *float64 `yaml:"weight"`
	PassScore *float64 `yaml:"pass_score"`
}

type evalSpec struct {
	Threshold *float64   `yaml:"threshold"`
	Cases     []caseSpec `yaml:"cases"`
}

// RunEval loads the eval set at evalSetPath, scores each case against target,
// and gates the weighted-mean aggregate on a threshold. A non-nil threshold
// overrides the eval set's own threshold (which defaults to 0.8 when absent).
// It is the "unit tests for prompts" engine: in CI a failing gate is meant to
// block the deploy.
func RunEval(evalSetPath string, target func(string) string, threshold *float64) (EvalReport, error) {
	data, err := os.ReadFile(evalSetPath)
	if err != nil {
		return EvalReport{}, err
	}
	var spec evalSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return EvalReport{}, err
	}

	effectiveThreshold := 0.8
	if spec.Threshold != nil {
		effectiveThreshold = *spec.Threshold
	}
	if threshold != nil {
		effectiveThreshold = *threshold
	}

	results := make([]CaseResult, 0, len(spec.Cases))
	for _, c := range spec.Cases {
		output := target(c.Input)
		scorer, err := GetScorer(c.Scorer)
		if err != nil {
			return EvalReport{}, err
		}
		arg := c.Expected
		if c.Scorer == "regex" {
			arg = c.Pattern
		}
		score := scorer(output, arg)
		weight := 1.0
		if c.Weight != nil {
			weight = *c.Weight
		}
		caseThreshold := 0.5
		if c.PassScore != nil {
			caseThreshold = *c.PassScore
		}
		results = append(results, CaseResult{
			ID:     c.ID,
			Scorer: c.Scorer,
			Weight: weight,
			Score:  score,
			Passed: score >= caseThreshold,
		})
	}

	totalW := 0.0
	for _, r := range results {
		totalW += r.Weight
	}
	if totalW == 0.0 {
		totalW = 1.0
	}
	weighted := 0.0
	for _, r := range results {
		weighted += r.Score * r.Weight
	}
	aggregate := round4(weighted / totalW)

	return EvalReport{
		Aggregate: aggregate,
		Threshold: effectiveThreshold,
		Passed:    aggregate >= effectiveThreshold,
		Cases:     results,
	}, nil
}

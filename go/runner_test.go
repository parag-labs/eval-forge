package evalforge

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func evalSetPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		candidate := filepath.Join(dir, "examples", "eval_set.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("examples/eval_set.yaml not found walking up from working dir")
		}
		dir = parent
	}
}

// writeTempEval writes an eval set into the package directory (never a system
// temp dir) and schedules its removal when the test finishes.
func writeTempEval(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(".", "eval-*.yaml")
	if err != nil {
		t.Fatalf("create temp eval: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp eval: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp eval: %v", err)
	}
	name := f.Name()
	t.Cleanup(func() { _ = os.Remove(name) })
	return name
}

func ptr(v float64) *float64 { return &v }

func TestGoodTargetPassesGate(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if !report.Passed {
		t.Fatalf("good target should pass the gate")
	}
	if report.Aggregate < report.Threshold {
		t.Fatalf("aggregate %v should meet threshold %v", report.Aggregate, report.Threshold)
	}
}

func TestAggregateExactValue(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if math.Abs(report.Aggregate-0.9238) > 1e-9 {
		t.Fatalf("aggregate = %v, want 0.9238", report.Aggregate)
	}
	if report.Threshold != 0.75 {
		t.Fatalf("threshold = %v, want 0.75", report.Threshold)
	}
}

func TestThreeCasesInEvalSetOrder(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if len(report.Cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(report.Cases))
	}
	wantIDs := []string{"refund-window", "cancellation", "no-hallucinated-discount"}
	for i, id := range wantIDs {
		if report.Cases[i].ID != id {
			t.Fatalf("case %d id = %q, want %q", i, report.Cases[i].ID, id)
		}
	}
}

func TestCaseWeightsAndScorers(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	wantWeight := []float64{2.0, 1.0, 1.5}
	wantScorer := []string{"contains", "semantic", "regex"}
	for i := range report.Cases {
		if report.Cases[i].Weight != wantWeight[i] {
			t.Fatalf("case %d weight = %v, want %v", i, report.Cases[i].Weight, wantWeight[i])
		}
		if report.Cases[i].Scorer != wantScorer[i] {
			t.Fatalf("case %d scorer = %q, want %q", i, report.Cases[i].Scorer, wantScorer[i])
		}
	}
}

func TestWeightsAffectAggregate(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	found := false
	for _, c := range report.Cases {
		if c.Weight == 2.0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a case with weight 2.0")
	}
	if report.Aggregate < 0.0 || report.Aggregate > 1.0 {
		t.Fatalf("aggregate %v out of [0,1]", report.Aggregate)
	}
}

func TestCaseScoresAndPassed(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if report.Cases[0].Score != 1.0 || !report.Cases[0].Passed {
		t.Fatalf("refund case = %+v, want score 1.0 passed", report.Cases[0])
	}
	if math.Abs(report.Cases[1].Score-0.6573) > 1e-9 || !report.Cases[1].Passed {
		t.Fatalf("cancellation case = %+v, want score 0.6573 passed", report.Cases[1])
	}
	if report.Cases[2].Score != 1.0 || !report.Cases[2].Passed {
		t.Fatalf("discount case = %+v, want score 1.0 passed", report.Cases[2])
	}
}

func TestRegressedTargetFailsGate(t *testing.T) {
	regressed := func(question string) string {
		if containsFold(question, "cancel") {
			return "Yes, cancel anytime."
		}
		return "I don't know."
	}
	report, err := RunEval(evalSetPath(t), regressed, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if report.Passed {
		t.Fatalf("regressed target should fail the gate")
	}
}

func containsFold(s, sub string) bool {
	return Contains(s, sub) == 1.0
}

func TestThresholdOverrideHigherFails(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, ptr(0.99))
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if report.Passed {
		t.Fatalf("aggregate should not clear an overridden 0.99 threshold")
	}
	if report.Threshold != 0.99 {
		t.Fatalf("threshold override = %v, want 0.99", report.Threshold)
	}
}

func TestThresholdOverrideLowerPasses(t *testing.T) {
	report, err := RunEval(evalSetPath(t), DemoTarget, ptr(0.1))
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if !report.Passed {
		t.Fatalf("aggregate should clear an overridden 0.1 threshold")
	}
	if report.Threshold != 0.1 {
		t.Fatalf("threshold override = %v, want 0.1", report.Threshold)
	}
}

func TestMissingFileErrors(t *testing.T) {
	if _, err := RunEval("does-not-exist.yaml", DemoTarget, nil); err == nil {
		t.Fatalf("expected an error for a missing eval set")
	}
}

func TestUnknownScorerErrors(t *testing.T) {
	path := writeTempEval(t, "threshold: 0.5\ncases:\n  - id: c1\n    input: hello\n    scorer: bogus\n    expected: hello\n")
	if _, err := RunEval(path, DemoTarget, nil); err == nil {
		t.Fatalf("expected an error for an unknown scorer")
	}
}

func TestZeroWeightsFallBackToOne(t *testing.T) {
	path := writeTempEval(t, "threshold: 0.0\ncases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n    weight: 0.0\n  - id: b\n    input: y\n    scorer: exact_match\n    expected: y\n    weight: 0.0\n")
	report, err := RunEval(path, DemoTarget, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if report.Aggregate != 0.0 {
		t.Fatalf("all-zero weights should aggregate to 0.0 without dividing by zero, got %v", report.Aggregate)
	}
	if !report.Passed {
		t.Fatalf("aggregate 0.0 should meet a 0.0 threshold")
	}
}

func TestDefaultThresholdWhenAbsent(t *testing.T) {
	path := writeTempEval(t, "cases:\n  - id: a\n    input: x\n    scorer: exact_match\n    expected: x\n")
	echo := func(s string) string { return s }
	report, err := RunEval(path, echo, nil)
	if err != nil {
		t.Fatalf("RunEval error: %v", err)
	}
	if report.Threshold != 0.8 {
		t.Fatalf("absent threshold should default to 0.8, got %v", report.Threshold)
	}
	if report.Aggregate != 1.0 {
		t.Fatalf("single passing case should aggregate to 1.0, got %v", report.Aggregate)
	}
}

func TestDemoTargetBranches(t *testing.T) {
	if DemoTarget("How many days for a refund?") != "You have 30 days from purchase to request a full refund." {
		t.Fatalf("refund branch wrong")
	}
	if DemoTarget("Can I cancel?") != "Yes, you can cancel your subscription at any time; it ends at the cycle close." {
		t.Fatalf("cancel branch wrong")
	}
	if DemoTarget("Any discount?") != "No, we do not offer a 90% loyalty discount." {
		t.Fatalf("discount branch wrong")
	}
	if DemoTarget("What is the weather?") != "I don't have information on that." {
		t.Fatalf("default branch wrong")
	}
}

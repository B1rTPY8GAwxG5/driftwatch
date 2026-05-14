package drift

import (
	"strings"
	"testing"
)

func cleanEvalResult() DriftResult {
	return DriftResult{Service: "api", Entries: nil}
}

func driftedEvalResult() DriftResult {
	return DriftResult{
		Service: "api",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		},
	}
}

func TestNewEvaluator_NotNil(t *testing.T) {
	e := NewEvaluator()
	if e == nil {
		t.Fatal("expected non-nil evaluator")
	}
}

func TestEvaluator_AddRule_NilConditionIgnored(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{Name: "bad", Condition: nil})
	out := e.Evaluate(cleanEvalResult())
	if len(out.Passed)+len(out.Failed) != 0 {
		t.Error("expected no rules to be applied")
	}
}

func TestEvaluator_Evaluate_AllPass(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{
		Name:      "no-drift",
		Condition: func(r DriftResult) bool { return !r.HasDrift() },
	})
	out := e.Evaluate(cleanEvalResult())
	if out.HasFailures() {
		t.Errorf("expected no failures, got %v", out.Failed)
	}
	if len(out.Passed) != 1 {
		t.Errorf("expected 1 passed rule, got %d", len(out.Passed))
	}
}

func TestEvaluator_Evaluate_RuleFails(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{
		Name:      "no-drift",
		Condition: func(r DriftResult) bool { return !r.HasDrift() },
	})
	out := e.Evaluate(driftedEvalResult())
	if !out.HasFailures() {
		t.Error("expected failure")
	}
	if out.Failed[0] != "no-drift" {
		t.Errorf("unexpected failed rule: %s", out.Failed[0])
	}
}

func TestEvaluationOutcome_Blocked_WhenFailed(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{
		Name:      "always-fail",
		Condition: func(r DriftResult) bool { return false },
	})
	out := e.Evaluate(cleanEvalResult())
	if !out.Blocked {
		t.Error("expected blocked=true when rule fails")
	}
}

func TestEvaluationOutcome_Summary_NoDrift(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{
		Name:      "check-clean",
		Condition: func(r DriftResult) bool { return true },
	})
	out := e.Evaluate(cleanEvalResult())
	if !strings.Contains(out.Summary(), "passed") {
		t.Errorf("expected 'passed' in summary, got: %s", out.Summary())
	}
}

func TestEvaluationOutcome_Summary_WithFailure(t *testing.T) {
	e := NewEvaluator()
	e.AddRule(EvaluationRule{
		Name:      "strict",
		Condition: func(r DriftResult) bool { return false },
	})
	out := e.Evaluate(driftedEvalResult())
	if !strings.Contains(out.Summary(), "failed") {
		t.Errorf("expected 'failed' in summary, got: %s", out.Summary())
	}
}

func TestEvaluator_EvaluateAll_Length(t *testing.T) {
	e := NewEvaluator()
	results := []DriftResult{cleanEvalResult(), driftedEvalResult()}
	outs := e.EvaluateAll(results)
	if len(outs) != 2 {
		t.Errorf("expected 2 outcomes, got %d", len(outs))
	}
}

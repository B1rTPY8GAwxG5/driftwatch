package drift

import (
	"testing"
)

var policyResult = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		{Kind: KindReplicas, Field: "replicas", Declared: "2", Observed: "3"},
	},
}

func TestPolicy_Evaluate_NoRules(t *testing.T) {
	p := &Policy{Name: "empty"}
	vs := p.Evaluate(policyResult)
	if len(vs) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(vs))
	}
}

func TestPolicy_Evaluate_MatchingRule(t *testing.T) {
	p := &Policy{
		Name: "strict",
		Rules: []PolicyRule{
			{Kind: KindImage, Action: PolicyActionBlock},
		},
	}
	vs := p.Evaluate(policyResult)
	if len(vs) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(vs))
	}
	if vs[0].Rule.Kind != KindImage {
		t.Errorf("expected KindImage, got %s", vs[0].Rule.Kind)
	}
}

func TestPolicy_Evaluate_IgnoreSkipped(t *testing.T) {
	p := &Policy{
		Name: "lenient",
		Rules: []PolicyRule{
			{Kind: KindImage, Action: PolicyActionIgnore},
			{Kind: KindReplicas, Action: PolicyActionWarn},
		},
	}
	vs := p.Evaluate(policyResult)
	if len(vs) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(vs))
	}
	if vs[0].Rule.Action != PolicyActionWarn {
		t.Errorf("expected warn action")
	}
}

func TestPolicy_Blocked_True(t *testing.T) {
	p := &Policy{}
	vs := []PolicyViolation{{Rule: PolicyRule{Kind: KindImage, Action: PolicyActionBlock}}}
	if !p.Blocked(vs) {
		t.Error("expected blocked=true")
	}
}

func TestPolicy_Blocked_False(t *testing.T) {
	p := &Policy{}
	vs := []PolicyViolation{{Rule: PolicyRule{Kind: KindImage, Action: PolicyActionWarn}}}
	if p.Blocked(vs) {
		t.Error("expected blocked=false")
	}
}

func TestPolicyViolation_Summary(t *testing.T) {
	v := PolicyViolation{
		Rule:    PolicyRule{Kind: KindImage, Action: PolicyActionBlock},
		Entries: []DriftEntry{{Field: "image"}},
	}
	s := v.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}

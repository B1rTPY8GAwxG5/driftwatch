package drift

import (
	"bytes"
	"strings"
	"testing"
)

var driftedForRemediation = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Want: "nginx:1.25", Got: "nginx:1.24"},
		{Kind: DriftKindReplicas, Field: "replicas", Want: 3, Got: 1},
		{Kind: DriftKindEnv, Field: "LOG_LEVEL", Want: "info", Got: "debug"},
	},
}

func TestBuildRemediationPlan_HasActions(t *testing.T) {
	plan := BuildRemediationPlan(driftedForRemediation)
	if !plan.HasActions() {
		t.Fatal("expected plan to have actions")
	}
	if len(plan.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(plan.Actions))
	}
}

func TestBuildRemediationPlan_NoActions(t *testing.T) {
	plan := BuildRemediationPlan(DriftResult{Service: "svc", Entries: nil})
	if plan.HasActions() {
		t.Fatal("expected no actions for clean result")
	}
}

func TestBuildRemediationPlan_ImageCommand(t *testing.T) {
	plan := BuildRemediationPlan(driftedForRemediation)
	if !strings.Contains(plan.Actions[0].Command, "kubectl set image") {
		t.Errorf("unexpected image command: %s", plan.Actions[0].Command)
	}
}

func TestBuildRemediationPlan_ReplicasCommand(t *testing.T) {
	plan := BuildRemediationPlan(driftedForRemediation)
	if !strings.Contains(plan.Actions[1].Command, "kubectl scale") {
		t.Errorf("unexpected replicas command: %s", plan.Actions[1].Command)
	}
}

func TestBuildRemediationPlan_EnvNoCommand(t *testing.T) {
	plan := BuildRemediationPlan(driftedForRemediation)
	if plan.Actions[2].Command != "" {
		t.Errorf("expected empty command for env drift, got %q", plan.Actions[2].Command)
	}
}

func TestRemediationPlan_WriteTo_NoDrift(t *testing.T) {
	plan := BuildRemediationPlan(DriftResult{Service: "svc"})
	var buf bytes.Buffer
	plan.WriteTo(&buf)
	if !strings.Contains(buf.String(), "No actions required") {
		t.Errorf("expected no-action message, got: %s", buf.String())
	}
}

func TestRemediationPlan_WriteTo_WithDrift(t *testing.T) {
	plan := BuildRemediationPlan(driftedForRemediation)
	var buf bytes.Buffer
	plan.WriteTo(&buf)
	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Error("expected service name in output")
	}
	if !strings.Contains(out, "kubectl set image") {
		t.Error("expected kubectl set image in output")
	}
}

package drift

import (
	"testing"
)

var labeledResult = DriftResult{
	Service: "api-gateway",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.23"},
	},
}

var cleanLabelResult = DriftResult{
	Service: "worker",
	Entries: []DriftEntry{},
}

func TestNewLabeler_NotNil(t *testing.T) {
	l := NewLabeler(nil)
	if l == nil {
		t.Fatal("expected non-nil Labeler")
	}
}

func TestLabeler_Label_StaticLabels(t *testing.T) {
	l := NewLabeler(map[string]string{"env": "prod", "team": "platform"})
	labels := l.Label(labeledResult)
	if labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", labels["env"])
	}
	if labels["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", labels["team"])
	}
}

func TestLabeler_Label_ServiceName(t *testing.T) {
	l := NewLabeler(nil)
	labels := l.Label(labeledResult)
	if labels["service"] != "api-gateway" {
		t.Errorf("expected service=api-gateway, got %q", labels["service"])
	}
}

func TestLabeler_Label_DriftedTrue(t *testing.T) {
	l := NewLabeler(nil)
	labels := l.Label(labeledResult)
	if labels["drifted"] != "true" {
		t.Errorf("expected drifted=true, got %q", labels["drifted"])
	}
}

func TestLabeler_Label_DriftedFalse(t *testing.T) {
	l := NewLabeler(nil)
	labels := l.Label(cleanLabelResult)
	if labels["drifted"] != "false" {
		t.Errorf("expected drifted=false, got %q", labels["drifted"])
	}
}

func TestLabeler_KindRule_Matches(t *testing.T) {
	l := NewLabeler(nil)
	l.AddKindRule(KindImage, "drift_type", "image")
	labels := l.Label(labeledResult)
	if labels["drift_type"] != "image" {
		t.Errorf("expected drift_type=image, got %q", labels["drift_type"])
	}
}

func TestLabeler_KindRule_NoMatch(t *testing.T) {
	l := NewLabeler(nil)
	l.AddKindRule(KindReplicas, "drift_type", "replicas")
	labels := l.Label(labeledResult)
	if _, ok := labels["drift_type"]; ok {
		t.Errorf("expected drift_type label to be absent")
	}
}

func TestLabelSet_String_NonEmpty(t *testing.T) {
	ls := LabelSet{"env": "staging"}
	s := ls.String()
	if s == "{}" {
		t.Error("expected non-empty string representation")
	}
}

func TestLabelSet_String_Empty(t *testing.T) {
	ls := LabelSet{}
	if ls.String() != "{}" {
		t.Errorf("expected {}, got %q", ls.String())
	}
}

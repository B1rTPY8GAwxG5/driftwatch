package drift

import (
	"testing"
)

func classifierImageEntry() DriftEntry {
	return DriftEntry{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.25"}
}

func classifierReplicasEntry() DriftEntry {
	return DriftEntry{Kind: KindReplicas, Field: "replicas", Declared: "3", Observed: "2"}
}

func TestNewClassifier_NotNil(t *testing.T) {
	c := NewClassifier(nil)
	if c == nil {
		t.Fatal("expected non-nil classifier")
	}
}

func TestClassifier_Classify_NoRules_Uncategorised(t *testing.T) {
	c := NewClassifier(nil)
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{classifierImageEntry()},
	}
	cr := c.Classify(result)
	if got := cr.Categories[entryKey(classifierImageEntry())]; got != "uncategorised" {
		t.Errorf("expected uncategorised, got %q", got)
	}
}

func TestClassifier_Classify_MatchingKind(t *testing.T) {
	rules := []ClassifierRule{
		{Kind: KindImage, Category: "container"},
	}
	c := NewClassifier(rules)
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{classifierImageEntry()},
	}
	cr := c.Classify(result)
	if got := cr.Categories[entryKey(classifierImageEntry())]; got != "container" {
		t.Errorf("expected container, got %q", got)
	}
}

func TestClassifier_Classify_MatchingKindAndField(t *testing.T) {
	rules := []ClassifierRule{
		{Kind: KindReplicas, Field: "replicas", Category: "scaling"},
	}
	c := NewClassifier(rules)
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{classifierReplicasEntry()},
	}
	cr := c.Classify(result)
	if got := cr.Categories[entryKey(classifierReplicasEntry())]; got != "scaling" {
		t.Errorf("expected scaling, got %q", got)
	}
}

func TestClassifier_Classify_FieldMismatch_Uncategorised(t *testing.T) {
	rules := []ClassifierRule{
		{Kind: KindReplicas, Field: "other", Category: "scaling"},
	}
	c := NewClassifier(rules)
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{classifierReplicasEntry()},
	}
	cr := c.Classify(result)
	if got := cr.Categories[entryKey(classifierReplicasEntry())]; got != "uncategorised" {
		t.Errorf("expected uncategorised, got %q", got)
	}
}

func TestClassifier_ClassifyAll_Length(t *testing.T) {
	c := NewClassifier(nil)
	results := []DriftResult{
		{Service: "a"},
		{Service: "b"},
	}
	out := c.ClassifyAll(results)
	if len(out) != 2 {
		t.Errorf("expected 2, got %d", len(out))
	}
}

func TestClassifier_Classify_PreservesResult(t *testing.T) {
	c := NewClassifier(nil)
	result := DriftResult{Service: "preserve-me"}
	cr := c.Classify(result)
	if cr.Service != "preserve-me" {
		t.Errorf("expected service name preserved, got %q", cr.Service)
	}
}

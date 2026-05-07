package drift

import (
	"errors"
	"testing"
)

func makePipeline(t *testing.T) (*Pipeline, ServiceSpec) {
	t.Helper()
	d := NewDetector()
	c := NewComparator(d)
	spec := ServiceSpec{Name: "svc", Image: "app:v1", Replicas: 2}
	return NewPipeline(c), spec
}

func TestNewPipeline_NotNil(t *testing.T) {
	p, _ := makePipeline(t)
	if p == nil {
		t.Fatal("expected non-nil Pipeline")
	}
}

func TestPipeline_Run_NoDrift(t *testing.T) {
	p, spec := makePipeline(t)
	result, err := p.Run(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasDrift() {
		t.Errorf("expected no drift")
	}
}

func TestPipeline_Run_StagesExecutedInOrder(t *testing.T) {
	p, spec := makePipeline(t)
	order := []int{}
	p.AddStage(func(r DriftResult) (DriftResult, error) {
		order = append(order, 1)
		return r, nil
	})
	p.AddStage(func(r DriftResult) (DriftResult, error) {
		order = append(order, 2)
		return r, nil
	})
	_, err := p.Run(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("unexpected stage order: %v", order)
	}
}

func TestPipeline_Run_StageError_Propagates(t *testing.T) {
	p, spec := makePipeline(t)
	expected := errors.New("stage error")
	p.AddStage(func(r DriftResult) (DriftResult, error) {
		return r, expected
	})
	_, err := p.Run(spec, spec)
	if !errors.Is(err, expected) {
		t.Errorf("expected stage error, got %v", err)
	}
}

func TestPipeline_ScoreStage_NoError(t *testing.T) {
	d := NewDetector()
	c := NewComparator(d)
	spec := ServiceSpec{Name: "svc", Image: "app:v1", Replicas: 2}
	live := ServiceSpec{Name: "svc", Image: "app:v2", Replicas: 2}
	p := NewPipeline(c, ScoreStage())
	_, err := p.Run(spec, live)
	if err != nil {
		t.Fatalf("unexpected error from ScoreStage: %v", err)
	}
}

func TestPipeline_LabelStage_AppliesLabels(t *testing.T) {
	d := NewDetector()
	c := NewComparator(d)
	spec := ServiceSpec{Name: "svc", Image: "app:v1", Replicas: 2}
	labeler := NewLabeler(map[string]string{"env": "prod"})
	p := NewPipeline(c, LabelStage(labeler))
	result, err := p.Run(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got %v", result.Labels)
	}
}

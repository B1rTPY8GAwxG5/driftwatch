package drift

import (
	"testing"
)

var scoredResult = DriftResult{
	Service: "scorer-svc",
	Entries: []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Expected: "v1", Actual: "v2"},
		{Kind: DriftKindEnv, Field: "ENV_KEY", Expected: "a", Actual: "b"},
	},
}

func TestScoreResult_CleanResult(t *testing.T) {
	result := DriftResult{Service: "clean-svc", Entries: nil}
	score := ScoreResult(result)
	if score.Total != 0 {
		t.Errorf("expected total 0, got %d", score.Total)
	}
	if score.Level != "clean" {
		t.Errorf("expected level clean, got %s", score.Level)
	}
}

func TestScoreResult_ImageDrift(t *testing.T) {
	result := DriftResult{
		Service: "img-svc",
		Entries: []DriftEntry{
			{Kind: DriftKindImage, Field: "image", Expected: "v1", Actual: "v2"},
		},
	}
	score := ScoreResult(result)
	if score.Total != 10 {
		t.Errorf("expected total 10, got %d", score.Total)
	}
	if score.Level != "critical" {
		t.Errorf("expected critical, got %s", score.Level)
	}
}

func TestScoreResult_MixedDrift(t *testing.T) {
	score := ScoreResult(scoredResult)
	expected := 13 // image(10) + env(3)
	if score.Total != expected {
		t.Errorf("expected total %d, got %d", expected, score.Total)
	}
	if score.Level != "critical" {
		t.Errorf("expected critical, got %s", score.Level)
	}
	if score.MaxEntry.Kind != DriftKindImage {
		t.Errorf("expected max entry kind image, got %s", score.MaxEntry.Kind)
	}
}

func TestScoreResult_ReplicasDrift(t *testing.T) {
	result := DriftResult{
		Service: "rep-svc",
		Entries: []DriftEntry{
			{Kind: DriftKindReplicas, Field: "replicas", Expected: "2", Actual: "3"},
		},
	}
	score := ScoreResult(result)
	if score.Total != 5 {
		t.Errorf("expected total 5, got %d", score.Total)
	}
	if score.Level != "warning" {
		t.Errorf("expected warning, got %s", score.Level)
	}
}

func TestScoreLevel_Info(t *testing.T) {
	level := scoreLevel(2)
	if level != "info" {
		t.Errorf("expected info, got %s", level)
	}
}

func TestScoreResult_ServiceName(t *testing.T) {
	score := ScoreResult(scoredResult)
	if score.Service != scoredResult.Service {
		t.Errorf("expected service %s, got %s", scoredResult.Service, score.Service)
	}
}

func TestScoreResult_EmptyEntries(t *testing.T) {
	result := DriftResult{Service: "empty-svc", Entries: []DriftEntry{}}
	score := ScoreResult(result)
	if score.Total != 0 {
		t.Errorf("expected total 0 for empty entries, got %d", score.Total)
	}
	if score.Level != "clean" {
		t.Errorf("expected level clean for empty entries, got %s", score.Level)
	}
}

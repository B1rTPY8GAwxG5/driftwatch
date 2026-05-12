package drift

import (
	"testing"
	"time"
)

var enrichBase = DriftResult{
	Service: "api-gateway",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
	},
}

func TestNewEnricher_NotNil(t *testing.T) {
	e := NewEnricher(EnrichmentContext{Environment: "prod"})
	if e == nil {
		t.Fatal("expected non-nil Enricher")
	}
}

func TestEnricher_Enrich_SetsContext(t *testing.T) {
	ctx := EnrichmentContext{
		Environment: "staging",
		Cluster:     "us-east-1",
		Region:      "us-east-1",
	}
	e := NewEnricher(ctx)
	er := e.Enrich(enrichBase)

	if er.Context.Environment != "staging" {
		t.Errorf("expected environment staging, got %s", er.Context.Environment)
	}
	if er.Context.Cluster != "us-east-1" {
		t.Errorf("expected cluster us-east-1, got %s", er.Context.Cluster)
	}
	if er.Result.Service != "api-gateway" {
		t.Errorf("expected service api-gateway, got %s", er.Result.Service)
	}
}

func TestEnricher_Enrich_SetsTimestamp(t *testing.T) {
	before := time.Now().UTC().Add(-time.Millisecond)
	e := NewEnricher(EnrichmentContext{})
	er := e.Enrich(enrichBase)
	after := time.Now().UTC().Add(time.Millisecond)

	if er.EnrichedAt.Before(before) || er.EnrichedAt.After(after) {
		t.Errorf("EnrichedAt %v out of expected range", er.EnrichedAt)
	}
}

func TestEnricher_EnrichAll_Length(t *testing.T) {
	e := NewEnricher(EnrichmentContext{Environment: "prod"})
	results := []DriftResult{enrichBase, enrichBase}
	out := e.EnrichAll(results)
	if len(out) != 2 {
		t.Errorf("expected 2 enriched results, got %d", len(out))
	}
}

func TestEnrichedResult_HasAnnotation_True(t *testing.T) {
	e := NewEnricher(EnrichmentContext{
		Annotations: map[string]string{"owner": "platform-team"},
	})
	er := e.Enrich(enrichBase)
	if !er.HasAnnotation("owner") {
		t.Error("expected HasAnnotation to return true for 'owner'")
	}
}

func TestEnrichedResult_HasAnnotation_False(t *testing.T) {
	e := NewEnricher(EnrichmentContext{})
	er := e.Enrich(enrichBase)
	if er.HasAnnotation("missing") {
		t.Error("expected HasAnnotation to return false for missing key")
	}
}

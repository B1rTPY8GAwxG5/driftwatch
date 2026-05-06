package drift

import (
	"testing"
	"time"
)

var (
	futureExpiry = time.Now().Add(24 * time.Hour)
	pastExpiry   = time.Now().Add(-24 * time.Hour)
)

func TestNewSuppressionStore_NotNil(t *testing.T) {
	s := NewSuppressionStore(nil)
	if s == nil {
		t.Fatal("expected non-nil SuppressionStore")
	}
}

func TestSuppressionRule_IsExpired_False(t *testing.T) {
	r := SuppressionRule{Expiry: futureExpiry}
	if r.IsExpired(time.Now()) {
		t.Error("expected rule to not be expired")
	}
}

func TestSuppressionRule_IsExpired_True(t *testing.T) {
	r := SuppressionRule{Expiry: pastExpiry}
	if !r.IsExpired(time.Now()) {
		t.Error("expected rule to be expired")
	}
}

func TestIsSuppressed_MatchingServiceAndKind(t *testing.T) {
	s := NewSuppressionStore([]SuppressionRule{
		{Service: "api", Kind: KindImage, Expiry: futureExpiry},
	})
	if !s.IsSuppressed("api", KindImage) {
		t.Error("expected drift to be suppressed")
	}
}

func TestIsSuppressed_WildcardKind(t *testing.T) {
	s := NewSuppressionStore([]SuppressionRule{
		{Service: "api", Kind: "", Expiry: futureExpiry},
	})
	if !s.IsSuppressed("api", KindReplicas) {
		t.Error("expected wildcard suppression to match any kind")
	}
}

func TestIsSuppressed_DifferentService(t *testing.T) {
	s := NewSuppressionStore([]SuppressionRule{
		{Service: "api", Kind: KindImage, Expiry: futureExpiry},
	})
	if s.IsSuppressed("worker", KindImage) {
		t.Error("expected different service to not be suppressed")
	}
}

func TestIsSuppressed_ExpiredRule(t *testing.T) {
	s := NewSuppressionStore([]SuppressionRule{
		{Service: "api", Kind: KindImage, Expiry: pastExpiry},
	})
	if s.IsSuppressed("api", KindImage) {
		t.Error("expected expired rule to not suppress")
	}
}

func TestPruneExpired_RemovesExpired(t *testing.T) {
	s := NewSuppressionStore([]SuppressionRule{
		{Service: "api", Kind: KindImage, Expiry: pastExpiry},
		{Service: "worker", Kind: KindReplicas, Expiry: futureExpiry},
	})
	s.PruneExpired()
	rules := s.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule after pruning, got %d", len(rules))
	}
	if rules[0].Service != "worker" {
		t.Errorf("expected remaining rule for 'worker', got '%s'", rules[0].Service)
	}
}

func TestAdd_AppendsRule(t *testing.T) {
	s := NewSuppressionStore(nil)
	s.Add(SuppressionRule{Service: "api", Kind: KindImage, Expiry: futureExpiry})
	if len(s.Rules()) != 1 {
		t.Errorf("expected 1 rule, got %d", len(s.Rules()))
	}
}

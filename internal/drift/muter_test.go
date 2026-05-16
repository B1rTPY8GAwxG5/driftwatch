package drift

import (
	"testing"
	"time"
)

func TestNewMuter_NotNil(t *testing.T) {
	m := NewMuter()
	if m == nil {
		t.Fatal("expected non-nil Muter")
	}
}

func TestMuter_IsMuted_NoRules_False(t *testing.T) {
	m := NewMuter()
	if m.IsMuted("svc-a", DriftKindImage) {
		t.Error("expected not muted with no rules")
	}
}

func TestMuter_IsMuted_ActiveRule_True(t *testing.T) {
	m := NewMuter()
	m.Add(MuteRule{
		Service:  "svc-a",
		Kind:     DriftKindImage,
		Deadline: time.Now().Add(time.Hour),
		Reason:   "planned maintenance",
	})
	if !m.IsMuted("svc-a", DriftKindImage) {
		t.Error("expected svc-a/image to be muted")
	}
}

func TestMuter_IsMuted_ExpiredRule_False(t *testing.T) {
	m := NewMuter()
	m.Add(MuteRule{
		Service:  "svc-a",
		Kind:     DriftKindImage,
		Deadline: time.Now().Add(-time.Minute),
	})
	if m.IsMuted("svc-a", DriftKindImage) {
		t.Error("expected expired rule to not mute")
	}
}

func TestMuter_IsMuted_WildcardKind_True(t *testing.T) {
	m := NewMuter()
	m.Add(MuteRule{
		Service:  "svc-b",
		Kind:     "",
		Deadline: time.Now().Add(time.Hour),
	})
	if !m.IsMuted("svc-b", DriftKindReplicas) {
		t.Error("expected wildcard kind to mute any drift kind")
	}
}

func TestMuter_Apply_RemovesMutedEntries(t *testing.T) {
	m := NewMuter()
	m.Add(MuteRule{
		Service:  "svc-a",
		Kind:     DriftKindImage,
		Deadline: time.Now().Add(time.Hour),
	})
	result := DriftResult{
		Service: "svc-a",
		Entries: []DriftEntry{
			{Kind: DriftKindImage, Field: "image"},
			{Kind: DriftKindReplicas, Field: "replicas"},
		},
	}
	out := m.Apply(result)
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry after muting, got %d", len(out.Entries))
	}
	if out.Entries[0].Kind != DriftKindReplicas {
		t.Errorf("expected remaining entry to be replicas, got %s", out.Entries[0].Kind)
	}
}

func TestMuter_Apply_NothingMuted_ReturnsAll(t *testing.T) {
	m := NewMuter()
	result := DriftResult{
		Service: "svc-c",
		Entries: []DriftEntry{
			{Kind: DriftKindImage, Field: "image"},
		},
	}
	out := m.Apply(result)
	if len(out.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(out.Entries))
	}
}

func TestMuter_Len_AfterAdd(t *testing.T) {
	m := NewMuter()
	m.Add(MuteRule{Service: "svc-x", Kind: DriftKindImage, Deadline: time.Now().Add(time.Hour)})
	m.Add(MuteRule{Service: "svc-y", Kind: DriftKindReplicas, Deadline: time.Now().Add(time.Hour)})
	if m.Len() != 2 {
		t.Errorf("expected 2 rules, got %d", m.Len())
	}
}

func TestMuteRule_IsExpired_False(t *testing.T) {
	r := MuteRule{Deadline: time.Now().Add(time.Hour)}
	if r.IsExpired(time.Now()) {
		t.Error("expected rule to not be expired")
	}
}

func TestMuteRule_IsExpired_True(t *testing.T) {
	r := MuteRule{Deadline: time.Now().Add(-time.Second)}
	if !r.IsExpired(time.Now()) {
		t.Error("expected rule to be expired")
	}
}

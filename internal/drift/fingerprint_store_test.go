package drift

import (
	"testing"
)

func makeStoreResult(service string) DriftResult {
	return DriftResult{
		Service: service,
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Detected: "nginx:1.25"},
		},
	}
}

func TestNewFingerprintStore_NotNil(t *testing.T) {
	s := NewFingerprintStore(nil)
	if s == nil {
		t.Fatal("expected non-nil FingerprintStore")
	}
}

func TestFingerprintStore_Record_FirstCall_Changed(t *testing.T) {
	s := NewFingerprintStore(nil)
	changed := s.Record(makeStoreResult("api"))
	if !changed {
		t.Fatal("expected changed=true on first record")
	}
}

func TestFingerprintStore_Record_SameResult_NotChanged(t *testing.T) {
	s := NewFingerprintStore(nil)
	r := makeStoreResult("api")
	s.Record(r)
	changed := s.Record(r)
	if changed {
		t.Fatal("expected changed=false for identical result")
	}
}

func TestFingerprintStore_Record_DifferentResult_Changed(t *testing.T) {
	s := NewFingerprintStore(nil)
	r1 := makeStoreResult("api")
	s.Record(r1)

	r2 := DriftResult{
		Service: "api",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Detected: "nginx:1.26"},
		},
	}
	changed := s.Record(r2)
	if !changed {
		t.Fatal("expected changed=true for different result")
	}
}

func TestFingerprintStore_Get_AfterRecord(t *testing.T) {
	s := NewFingerprintStore(nil)
	r := makeStoreResult("api")
	s.Record(r)
	_, ok := s.Get("api")
	if !ok {
		t.Fatal("expected fingerprint to be present after Record")
	}
}

func TestFingerprintStore_Get_Missing(t *testing.T) {
	s := NewFingerprintStore(nil)
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected missing fingerprint for unknown service")
	}
}

func TestFingerprintStore_Delete_RemovesEntry(t *testing.T) {
	s := NewFingerprintStore(nil)
	s.Record(makeStoreResult("api"))
	s.Delete("api")
	_, ok := s.Get("api")
	if ok {
		t.Fatal("expected fingerprint to be removed after Delete")
	}
}

func TestFingerprintStore_Flush_ClearsAll(t *testing.T) {
	s := NewFingerprintStore(nil)
	s.Record(makeStoreResult("api"))
	s.Record(makeStoreResult("worker"))
	s.Flush()
	if s.Len() != 0 {
		t.Fatalf("expected 0 entries after Flush, got %d", s.Len())
	}
}

func TestFingerprintStore_Len(t *testing.T) {
	s := NewFingerprintStore(nil)
	s.Record(makeStoreResult("api"))
	s.Record(makeStoreResult("worker"))
	if s.Len() != 2 {
		t.Fatalf("expected Len=2, got %d", s.Len())
	}
}

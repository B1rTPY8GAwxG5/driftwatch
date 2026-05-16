package drift

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func archiverResult(service string) DriftResult {
	return DriftResult{
		Service: service,
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.25"},
		},
	}
}

func TestDefaultArchivePolicy_Values(t *testing.T) {
	p := DefaultArchivePolicy()
	if p.MaxAge != 30*24*time.Hour {
		t.Errorf("expected 30d, got %v", p.MaxAge)
	}
	if p.Dir == "" {
		t.Error("expected non-empty Dir")
	}
}

func TestNewArchiver_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	policy := ArchivePolicy{Dir: filepath.Join(dir, "archive"), MaxAge: time.Hour}
	a, err := NewArchiver(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil archiver")
	}
	if _, err := os.Stat(policy.Dir); err != nil {
		t.Errorf("expected dir to exist: %v", err)
	}
}

func TestArchiver_Archive_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewArchiver(ArchivePolicy{Dir: dir, MaxAge: time.Hour})
	if err := a.Archive(archiverResult("svc-a")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, _ := a.Count()
	if n != 1 {
		t.Errorf("expected 1 file, got %d", n)
	}
}

func TestArchiver_Archive_MultipleResults(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewArchiver(ArchivePolicy{Dir: dir, MaxAge: time.Hour})
	for i := 0; i < 3; i++ {
		time.Sleep(time.Millisecond) // ensure unique timestamps
		_ = a.Archive(archiverResult("svc-b"))
	}
	n, _ := a.Count()
	if n != 3 {
		t.Errorf("expected 3 files, got %d", n)
	}
}

func TestArchiver_Prune_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewArchiver(ArchivePolicy{Dir: dir, MaxAge: time.Millisecond})
	_ = a.Archive(archiverResult("svc-c"))
	time.Sleep(5 * time.Millisecond)
	removed, err := a.Prune()
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
	n, _ := a.Count()
	if n != 0 {
		t.Errorf("expected 0 files after prune, got %d", n)
	}
}

func TestArchiver_Prune_KeepsRecentFiles(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewArchiver(ArchivePolicy{Dir: dir, MaxAge: time.Hour})
	_ = a.Archive(archiverResult("svc-d"))
	removed, err := a.Prune()
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestNewArchiver_ZeroPolicy_UsesDefaults(t *testing.T) {
	dir := t.TempDir()
	a, err := NewArchiver(ArchivePolicy{Dir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.policy.MaxAge != DefaultArchivePolicy().MaxAge {
		t.Errorf("expected default MaxAge, got %v", a.policy.MaxAge)
	}
}

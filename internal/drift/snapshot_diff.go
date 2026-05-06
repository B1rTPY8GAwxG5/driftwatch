package drift

import "fmt"

// SnapshotDiff describes what changed between two snapshots of the same service.
type SnapshotDiff struct {
	Service  string
	Resolved []DriftEntry // entries present in previous but absent in current
	New      []DriftEntry // entries absent in previous but present in current
}

// HasChanges reports whether the diff contains any resolved or new entries.
func (sd SnapshotDiff) HasChanges() bool {
	return len(sd.Resolved) > 0 || len(sd.New) > 0
}

// Summary returns a human-readable summary of the snapshot diff.
func (sd SnapshotDiff) Summary() string {
	if !sd.HasChanges() {
		return fmt.Sprintf("[%s] no changes since last snapshot", sd.Service)
	}
	return fmt.Sprintf("[%s] %d resolved, %d new drift entries",
		sd.Service, len(sd.Resolved), len(sd.New))
}

// CompareSnapshots computes the diff between a previous and current DriftResult.
func CompareSnapshots(prev, curr DriftResult) SnapshotDiff {
	sd := SnapshotDiff{Service: curr.Service}

	prevIndex := indexEntries(prev.Entries)
	currIndex := indexEntries(curr.Entries)

	for key, entry := range prevIndex {
		if _, found := currIndex[key]; !found {
			sd.Resolved = append(sd.Resolved, entry)
		}
	}
	for key, entry := range currIndex {
		if _, found := prevIndex[key]; !found {
			sd.New = append(sd.New, entry)
		}
	}
	return sd
}

func indexEntries(entries []DriftEntry) map[string]DriftEntry {
	m := make(map[string]DriftEntry, len(entries))
	for _, e := range entries {
		key := fmt.Sprintf("%s::%s", e.Kind, e.Field)
		m[key] = e
	}
	return m
}

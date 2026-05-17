package drift

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// IndexMode controls how results are indexed.
type IndexMode string

const (
	IndexByService IndexMode = "service"
	IndexByKind    IndexMode = "kind"
	IndexByBoth    IndexMode = "both"
)

// DriftIndex provides fast lookup of DriftResults by key.
type DriftIndex struct {
	mu      sync.RWMutex
	mode    IndexMode
	buckets map[string][]DriftResult
}

// NewIndexer creates a new DriftIndex with the given mode.
// Unknown modes fall back to IndexByService.
func NewIndexer(mode IndexMode) *DriftIndex {
	switch mode {
	case IndexByService, IndexByKind, IndexByBoth:
	default:
		mode = IndexByService
	}
	return &DriftIndex{
		mode:    mode,
		buckets: make(map[string][]DriftResult),
	}
}

// Index stores a DriftResult under its computed key(s).
func (idx *DriftIndex) Index(r DriftResult) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	for _, k := range idx.keysFor(r) {
		idx.buckets[k] = append(idx.buckets[k], r)
	}
}

// Lookup returns all results stored under the given key.
func (idx *DriftIndex) Lookup(key string) []DriftResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]DriftResult, len(idx.buckets[key]))
	copy(out, idx.buckets[key])
	return out
}

// Keys returns all indexed keys in sorted order.
func (idx *DriftIndex) Keys() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	keys := make([]string, 0, len(idx.buckets))
	for k := range idx.buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Flush removes all indexed data.
func (idx *DriftIndex) Flush() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.buckets = make(map[string][]DriftResult)
}

func (idx *DriftIndex) keysFor(r DriftResult) []string {
	switch idx.mode {
	case IndexByKind:
		kinds := kindSet(r)
		out := make([]string, 0, len(kinds))
		for k := range kinds {
			out = append(out, string(k))
		}
		return out
	case IndexByBoth:
		kinds := kindSet(r)
		out := make([]string, 0, len(kinds)+1)
		out = append(out, r.Service)
		for k := range kinds {
			out = append(out, fmt.Sprintf("%s::%s", r.Service, k))
		}
		return out
	default:
		return []string{r.Service}
	}
}

func kindSet(r DriftResult) map[DriftKind]struct{} {
	kinds := make(map[DriftKind]struct{})
	for _, e := range r.Entries {
		kinds[e.Kind] = struct{}{}
	}
	if len(kinds) == 0 {
		kinds[DriftKind("none")] = struct{}{}
	}
	return kinds
}

// Mode returns the current index mode as a string.
func (idx *DriftIndex) Mode() string {
	return strings.ToLower(string(idx.mode))
}

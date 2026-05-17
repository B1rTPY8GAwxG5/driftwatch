package drift

import "fmt"

// MapMode controls how results are keyed in the output map.
type MapMode string

const (
	MapByService MapMode = "service"
	MapByKind    MapMode = "kind"
)

// Mapper organises a slice of DriftResults into a keyed map for fast lookup.
type Mapper struct {
	mode MapMode
}

// NewMapper returns a Mapper using the given mode. Unknown modes default to MapByService.
func NewMapper(mode MapMode) *Mapper {
	switch mode {
	case MapByService, MapByKind:
	default:
		mode = MapByService
	}
	return &Mapper{mode: mode}
}

// Mode returns the active mapping mode.
func (m *Mapper) Mode() MapMode { return m.mode }

// Map converts a slice of DriftResult into a map keyed by the configured mode.
// When mode is MapByKind each key may aggregate multiple results.
func (m *Mapper) Map(results []DriftResult) map[string][]DriftResult {
	out := make(map[string][]DriftResult)
	for _, r := range results {
		keys := m.keysFor(r)
		for _, k := range keys {
			out[k] = append(out[k], r)
		}
	}
	return out
}

// Lookup returns the results associated with the given key, or nil if absent.
func (m *Mapper) Lookup(results []DriftResult, key string) []DriftResult {
	return m.Map(results)[key]
}

func (m *Mapper) keysFor(r DriftResult) []string {
	switch m.mode {
	case MapByKind:
		if len(r.Entries) == 0 {
			return []string{"none"}
		}
		seen := make(map[string]struct{})
		var keys []string
		for _, e := range r.Entries {
			k := fmt.Sprintf("%s", e.Kind)
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				keys = append(keys, k)
			}
		}
		return keys
	default:
		return []string{r.Service}
	}
}

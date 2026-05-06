package drift

// SeverityScore maps drift kinds to numeric severity weights.
var SeverityScore = map[DriftKind]int{
	DriftKindImage:    10,
	DriftKindReplicas: 5,
	DriftKindEnv:      3,
}

// DriftScore holds the computed score for a drift result.
type DriftScore struct {
	Service  string
	Total    int
	MaxEntry DriftEntry
	Level    string
}

// ScoreResult computes a weighted severity score for a DriftResult.
func ScoreResult(result DriftResult) DriftScore {
	if !result.HasDrift() {
		return DriftScore{Service: result.Service, Total: 0, Level: "clean"}
	}

	total := 0
	var maxEntry DriftEntry
	maxScore := -1

	for _, entry := range result.Entries {
		w, ok := SeverityScore[entry.Kind]
		if !ok {
			w = 1
		}
		total += w
		if w > maxScore {
			maxScore = w
			maxEntry = entry
		}
	}

	return DriftScore{
		Service:  result.Service,
		Total:    total,
		MaxEntry: maxEntry,
		Level:    scoreLevel(total),
	}
}

// scoreLevel returns a human-readable severity level based on total score.
func scoreLevel(total int) string {
	switch {
	case total >= 10:
		return "critical"
	case total >= 5:
		return "warning"
	case total > 0:
		return "info"
	default:
		return "clean"
	}
}

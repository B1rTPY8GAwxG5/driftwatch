package drift

import (
	"math"
	"sync"
	"time"
)

// DecayModel tracks a drift score that decays over time when no drift is observed.
type DecayModel struct {
	mu        sync.Mutex
	scores    map[string]float64
	lastSeen  map[string]time.Time
	halfLife  time.Duration
	floor     float64
}

// NewDecayModel creates a DecayModel with the given half-life duration.
// The score halves every halfLife duration when no new drift is recorded.
// floor is the minimum score value (clamped above zero).
func NewDecayModel(halfLife time.Duration, floor float64) *DecayModel {
	if halfLife <= 0 {
		halfLife = 10 * time.Minute
	}
	if floor < 0 {
		floor = 0
	}
	return &DecayModel{
		scores:   make(map[string]float64),
		lastSeen: make(map[string]time.Time),
		halfLife: halfLife,
		floor:    floor,
	}
}

// Record adds delta to the score for the given service key and resets its decay clock.
func (d *DecayModel) Record(service string, delta float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	current := d.decayed(service, time.Now())
	d.scores[service] = current + delta
	d.lastSeen[service] = time.Now()
}

// Score returns the current decayed score for the service.
func (d *DecayModel) Score(service string) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.decayed(service, time.Now())
}

// Reset clears the score for the given service.
func (d *DecayModel) Reset(service string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.scores, service)
	delete(d.lastSeen, service)
}

// decayed computes the exponentially decayed score at the given time (caller holds lock).
func (d *DecayModel) decayed(service string, now time.Time) float64 {
	score, ok := d.scores[service]
	if !ok {
		return d.floor
	}
	last, ok := d.lastSeen[service]
	if !ok {
		return d.floor
	}
	elapsed := now.Sub(last)
	decayFactor := math.Pow(0.5, elapsed.Seconds()/d.halfLife.Seconds())
	result := score * decayFactor
	if result < d.floor {
		return d.floor
	}
	return result
}

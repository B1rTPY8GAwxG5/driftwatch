package drift

import "fmt"

// SignalEmitter wraps a SignalBus and produces typed signals from DriftResults.
type SignalEmitter struct {
	bus    *SignalBus
	labels map[string]string
}

// NewSignalEmitter creates a SignalEmitter backed by the given bus.
func NewSignalEmitter(bus *SignalBus, staticLabels map[string]string) *SignalEmitter {
	if bus == nil {
		bus = NewSignalBus()
	}
	lbls := make(map[string]string)
	for k, v := range staticLabels {
		lbls[k] = v
	}
	return &SignalEmitter{bus: bus, labels: lbls}
}

// Emit inspects a DriftResult and publishes an appropriate signal.
func (e *SignalEmitter) Emit(result DriftResult) {
	kind := SignalDriftResolved
	msg := fmt.Sprintf("service %q is clean", result.Service)
	if result.HasDrift() {
		kind = SignalDriftDetected
		msg = fmt.Sprintf("service %q has %d drifted field(s)", result.Service, len(result.Entries))
	}
	e.bus.Publish(Signal{
		Kind:    kind,
		Service: result.Service,
		Message: msg,
		Labels:  e.mergedLabels(result),
	})
}

// EmitBudgetExceeded publishes a budget-exceeded signal for the given service.
func (e *SignalEmitter) EmitBudgetExceeded(service string, used, limit int) {
	e.bus.Publish(Signal{
		Kind:    SignalBudgetExceeded,
		Service: service,
		Message: fmt.Sprintf("drift budget exceeded: %d/%d", used, limit),
		Labels:  e.labels,
	})
}

func (e *SignalEmitter) mergedLabels(result DriftResult) map[string]string {
	out := make(map[string]string, len(e.labels)+1)
	for k, v := range e.labels {
		out[k] = v
	}
	out["service"] = result.Service
	return out
}

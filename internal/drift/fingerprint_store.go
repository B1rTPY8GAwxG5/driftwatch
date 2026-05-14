package drift

import (
	"fmt"
	"sync"
)

// FingerprintStore tracks the last known fingerprint per service.
// It can detect whether a drift result represents a change from the
// previously recorded state.
type FingerprintStore struct {
	mu           sync.RWMutex
	fingerprints map[string]Fingerprint
	printer      *Fingerprinter
}

// NewFingerprintStore creates a FingerprintStore using the given Fingerprinter.
func NewFingerprintStore(printer *Fingerprinter) *FingerprintStore {
	if printer == nil {
		printer = NewFingerprinter(DefaultFingerprintOptions())
	}
	return &FingerprintStore{
		fingerprints: make(map[string]Fingerprint),
		printer:      printer,
	}
}

// Record stores the fingerprint for the given result and returns whether
// the fingerprint has changed since the last call for the same service.
func (s *FingerprintStore) Record(r DriftResult) (changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.printer.Compute(r)
	prev, exists := s.fingerprints[r.Service]
	s.fingerprints[r.Service] = next

	if !exists {
		return true
	}
	return !prev.Equal(next)
}

// Get returns the last recorded fingerprint for the service.
func (s *FingerprintStore) Get(service string) (Fingerprint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fp, ok := s.fingerprints[service]
	return fp, ok
}

// Delete removes the stored fingerprint for the given service.
func (s *FingerprintStore) Delete(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.fingerprints, service)
}

// Flush removes all stored fingerprints.
func (s *FingerprintStore) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fingerprints = make(map[string]Fingerprint)
}

// Len returns the number of tracked services.
func (s *FingerprintStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.fingerprints)
}

// String returns a human-readable summary of the store.
func (s *FingerprintStore) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("FingerprintStore{services: %d}", len(s.fingerprints))
}

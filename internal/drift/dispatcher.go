package drift

import (
	"fmt"
	"io"
	"sync"
)

// DispatchMode controls how results are dispatched to handlers.
type DispatchMode int

const (
	DispatchSerial   DispatchMode = iota // handlers called sequentially
	DispatchParallel                     // handlers called concurrently
)

// DispatchHandler is a function that receives a DriftResult.
type DispatchHandler func(result DriftResult) error

// Dispatcher fans out DriftResults to multiple registered handlers.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers []DispatchHandler
	mode     DispatchMode
	errWriter io.Writer
}

// NewDispatcher creates a Dispatcher with the given mode and error writer.
func NewDispatcher(mode DispatchMode, errWriter io.Writer) *Dispatcher {
	if errWriter == nil {
		errWriter = io.Discard
	}
	return &Dispatcher{mode: mode, errWriter: errWriter}
}

// Register adds a handler to the dispatcher.
func (d *Dispatcher) Register(h DispatchHandler) {
	if h == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, h)
}

// Dispatch sends the result to all registered handlers.
func (d *Dispatcher) Dispatch(result DriftResult) {
	d.mu.RLock()
	handlers := make([]DispatchHandler, len(d.handlers))
	copy(handlers, d.handlers)
	d.mu.RUnlock()

	if d.mode == DispatchParallel {
		var wg sync.WaitGroup
		for _, h := range handlers {
			wg.Add(1)
			go func(fn DispatchHandler) {
				defer wg.Done()
				if err := fn(result); err != nil {
					fmt.Fprintf(d.errWriter, "dispatcher: handler error: %v\n", err)
				}
			}(h)
		}
		wg.Wait()
		return
	}

	for _, h := range handlers {
		if err := h(result); err != nil {
			fmt.Fprintf(d.errWriter, "dispatcher: handler error: %v\n", err)
		}
	}
}

// Len returns the number of registered handlers.
func (d *Dispatcher) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}

package drift

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
)

func TestNewDispatcher_NotNil(t *testing.T) {
	d := NewDispatcher(DispatchSerial, nil)
	if d == nil {
		t.Fatal("expected non-nil Dispatcher")
	}
}

func TestDispatcher_Register_NilHandlerIgnored(t *testing.T) {
	d := NewDispatcher(DispatchSerial, nil)
	d.Register(nil)
	if d.Len() != 0 {
		t.Errorf("expected 0 handlers, got %d", d.Len())
	}
}

func TestDispatcher_Len_AfterRegister(t *testing.T) {
	d := NewDispatcher(DispatchSerial, nil)
	d.Register(func(DriftResult) error { return nil })
	d.Register(func(DriftResult) error { return nil })
	if d.Len() != 2 {
		t.Errorf("expected 2, got %d", d.Len())
	}
}

func TestDispatcher_Serial_AllHandlersCalled(t *testing.T) {
	d := NewDispatcher(DispatchSerial, nil)
	var count int32
	for i := 0; i < 3; i++ {
		d.Register(func(DriftResult) error {
			atomic.AddInt32(&count, 1)
			return nil
		})
	}
	d.Dispatch(DriftResult{Service: "svc"})
	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 calls, got %d", count)
	}
}

func TestDispatcher_Parallel_AllHandlersCalled(t *testing.T) {
	d := NewDispatcher(DispatchParallel, nil)
	var count int32
	for i := 0; i < 5; i++ {
		d.Register(func(DriftResult) error {
			atomic.AddInt32(&count, 1)
			return nil
		})
	}
	d.Dispatch(DriftResult{Service: "svc"})
	if atomic.LoadInt32(&count) != 5 {
		t.Errorf("expected 5 calls, got %d", count)
	}
}

func TestDispatcher_HandlerError_WritesToErrWriter(t *testing.T) {
	var buf bytes.Buffer
	d := NewDispatcher(DispatchSerial, &buf)
	d.Register(func(DriftResult) error {
		return errors.New("handler failed")
	})
	d.Dispatch(DriftResult{Service: "svc"})
	if buf.Len() == 0 {
		t.Error("expected error written to errWriter")
	}
}

func TestDispatchMode_Constants(t *testing.T) {
	if DispatchSerial == DispatchParallel {
		t.Error("DispatchSerial and DispatchParallel must differ")
	}
}

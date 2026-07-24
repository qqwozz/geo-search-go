package circuitbreaker

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StartsClosed(t *testing.T) {
	cb := New(3, 2, time.Second)
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %d", cb.State())
	}
	if !cb.Allow() {
		t.Error("should allow when closed")
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := New(3, 2, time.Second)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen, got %d", cb.State())
	}
	if cb.Allow() {
		t.Error("should not allow when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := New(3, 2, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(100 * time.Millisecond)
	if !cb.Allow() {
		t.Error("should allow after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %d", cb.State())
	}
}

func TestCircuitBreaker_ClosesAfterSuccesses(t *testing.T) {
	cb := New(3, 2, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(100 * time.Millisecond)
	cb.Allow() // transitions to half-open
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %d", cb.State())
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	cb := New(3, 2, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(100 * time.Millisecond)
	cb.Allow() // half-open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after failure in half-open, got %d", cb.State())
	}
}

func TestCircuitBreaker_ResetsOnSuccessInClosed(t *testing.T) {
	cb := New(3, 2, time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // resets
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %d", cb.State())
	}
}

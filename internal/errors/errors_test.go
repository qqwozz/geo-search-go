package errors

import (
	"net/http"
	"testing"
)

func TestAppError(t *testing.T) {
	err := New(http.StatusBadRequest, "invalid input")
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode())
	}
	if err.Error() != "invalid input" {
		t.Errorf("expected 'invalid input', got '%s'", err.Error())
	}
}

func TestAppErrorf(t *testing.T) {
	err := Newf(http.StatusNotFound, "user %d not found", 42)
	if err.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode())
	}
	if err.Error() != "user 42 not found" {
		t.Errorf("expected 'user 42 not found', got '%s'", err.Error())
	}
}

func TestAppErrorWrap(t *testing.T) {
	inner := New(http.StatusInternalServerError, "db error")
	err := Wrap(http.StatusInternalServerError, "search failed", inner)
	if err.Detail != "db error" {
		t.Errorf("expected detail 'db error', got '%s'", err.Detail)
	}
	if err.Error() != "search failed: db error" {
		t.Errorf("expected error 'search failed: db error', got '%s'", err.Error())
	}
}

func TestPredefinedErrors(t *testing.T) {
	err := ErrBadRequest("bad query")
	if err.StatusCode() != http.StatusBadRequest {
		t.Error("ErrBadRequest should be 400")
	}

	err = ErrNotFound("not found")
	if err.StatusCode() != http.StatusNotFound {
		t.Error("ErrNotFound should be 404")
	}

	err = ErrInternal("internal")
	if err.StatusCode() != http.StatusInternalServerError {
		t.Error("ErrInternal should be 500")
	}

	err = ErrTooManyRequests()
	if err.StatusCode() != http.StatusTooManyRequests {
		t.Error("ErrTooManyRequests should be 429")
	}

	err = ErrServiceTimeout()
	if err.StatusCode() != http.StatusServiceUnavailable {
		t.Error("ErrServiceTimeout should be 503")
	}
}

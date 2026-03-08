package usecase

import (
	"errors"
	"testing"
)

func TestAppErrorFormat(t *testing.T) {
	err := NewAppError("ERR_HAS_HOLDINGS", ErrHasHoldings)
	want := "ERR_HAS_HOLDINGS|portfolio has holdings"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAppErrorFields(t *testing.T) {
	err := NewAppError("ERR_EMPTY_NAME", ErrEmptyName)
	if err.Code != "ERR_EMPTY_NAME" {
		t.Errorf("Code = %q, want ERR_EMPTY_NAME", err.Code)
	}
	if err.Message != "name is required" {
		t.Errorf("Message = %q, want %q", err.Message, "name is required")
	}
}

func TestNewAppErrorPreservesMessage(t *testing.T) {
	sentinel := errors.New("custom error")
	appErr := NewAppError("ERR_CUSTOM", sentinel)
	if appErr.Message != "custom error" {
		t.Errorf("Message = %q, want %q", appErr.Message, "custom error")
	}
}

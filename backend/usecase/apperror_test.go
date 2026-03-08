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
	if err.Err.Error() != ErrEmptyName.Error() {
		t.Errorf("Err.Error() = %q, want %q", err.Err.Error(), ErrEmptyName.Error())
	}
}

func TestNewAppErrorPreservesMessage(t *testing.T) {
	sentinel := errors.New("custom error")
	appErr := NewAppError("ERR_CUSTOM", sentinel)
	if appErr.Err.Error() != "custom error" {
		t.Errorf("Err.Error() = %q, want %q", appErr.Err.Error(), "custom error")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	appErr := NewAppError("ERR_HAS_HOLDINGS", ErrHasHoldings)
	if !errors.Is(appErr, ErrHasHoldings) {
		t.Error("errors.Is should match the wrapped sentinel")
	}
}

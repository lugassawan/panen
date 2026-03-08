package presenter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lugassawan/panen/backend/usecase"
)

func TestToAppErrorNil(t *testing.T) {
	if toAppError(nil) != nil {
		t.Error("toAppError(nil) should return nil")
	}
}

func TestToAppErrorKnownSentinel(t *testing.T) {
	err := toAppError(usecase.ErrHasHoldings)
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != "ERR_HAS_HOLDINGS" {
		t.Errorf("Code = %q, want ERR_HAS_HOLDINGS", appErr.Code)
	}
}

func TestToAppErrorWrappedSentinel(t *testing.T) {
	wrapped := fmt.Errorf("delete portfolio: %w", usecase.ErrHasHoldings)
	err := toAppError(wrapped)
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != "ERR_HAS_HOLDINGS" {
		t.Errorf("Code = %q, want ERR_HAS_HOLDINGS", appErr.Code)
	}
}

func TestToAppErrorUnknown(t *testing.T) {
	unknown := errors.New("something unexpected")
	err := toAppError(unknown)
	var appErr *usecase.AppError
	if errors.As(err, &appErr) {
		t.Error("unknown errors should not be converted to AppError")
	}
	if !errors.Is(err, unknown) {
		t.Error("unknown errors should pass through unchanged")
	}
}

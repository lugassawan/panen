package presenter

import (
	"errors"

	"github.com/lugassawan/panen/backend/usecase"
)

// errorMapping maps sentinel errors to structured error codes.
var errorMapping = []struct {
	sentinel error
	code     string
}{
	{usecase.ErrHasHoldings, "ERR_HAS_HOLDINGS"},
	{usecase.ErrHasDependents, "ERR_HAS_DEPENDENTS"},
	{usecase.ErrDuplicateMode, "ERR_DUPLICATE_MODE"},
	{usecase.ErrDuplicateHolding, "ERR_DUPLICATE_HOLDING"},
	{usecase.ErrModeImmutable, "ERR_MODE_IMMUTABLE"},
	{usecase.ErrEmptyName, "ERR_EMPTY_NAME"},
	{usecase.ErrEmptyTicker, "ERR_EMPTY_TICKER"},
	{usecase.ErrInvalidPrice, "ERR_INVALID_PRICE"},
	{usecase.ErrInvalidLots, "ERR_INVALID_LOTS"},
	{usecase.ErrWatchlistNameTaken, "ERR_WATCHLIST_NAME_TAKEN"},
}

// toAppError converts a known sentinel error to an AppError with a structured
// code. Unknown errors pass through unchanged.
func toAppError(err error) error {
	if err == nil {
		return nil
	}
	for _, m := range errorMapping {
		if errors.Is(err, m.sentinel) {
			return usecase.NewAppError(m.code, err)
		}
	}
	return err
}

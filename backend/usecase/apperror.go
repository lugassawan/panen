package usecase

// AppError is a structured error carrying a machine-readable code for frontend i18n.
// Format: "CODE|human-readable message".
type AppError struct {
	Code string
	Err  error
}

// NewAppError creates a new AppError from a code and an existing error.
func NewAppError(code string, err error) *AppError {
	return &AppError{Code: code, Err: err}
}

func (e *AppError) Error() string { return e.Code + "|" + e.Err.Error() }
func (e *AppError) Unwrap() error { return e.Err }

package registry

import "errors"

var (
	ErrInvalidModelID       = errors.New("invalid model ID")
	ErrInvalidLabel         = errors.New("invalid model label")
	ErrInvalidTableName     = errors.New("invalid model table name")
	ErrInvalidVersion       = errors.New("invalid model version")
	ErrMissingOwnerModule   = errors.New("missing model owner module")
	ErrInvalidModuleID      = errors.New("invalid module ID")
	ErrInvalidSource        = errors.New("invalid model source")
	ErrDuplicateModel       = errors.New("model already registered")
	ErrModelVersionConflict = errors.New("model version conflict")
	ErrUnknownModel         = errors.New("unknown model")
)

// RegistrationError provides a stable error category and model context while
// preserving errors.Is compatibility with the package sentinel errors.
type RegistrationError struct {
	ModelID ModelID
	Cause   error
	Detail  string
}

func (e *RegistrationError) Error() string {
	if e.Detail == "" {
		return e.Cause.Error() + ": " + string(e.ModelID)
	}
	return e.Cause.Error() + ": " + string(e.ModelID) + ": " + e.Detail
}

func (e *RegistrationError) Unwrap() error {
	return e.Cause
}

package servers

import "errors"

var (
	ErrSessionInvalidated = errors.New("session has been invalidated")
	ErrSessionExpired     = errors.New("session has expired")
)

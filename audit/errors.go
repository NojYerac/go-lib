package audit

import "errors"

var (
	ErrInvalidEvent     = errors.New("invalid audit event")
	ErrInvalidAction    = errors.New("invalid audit action")
	ErrInvalidActor     = errors.New("invalid audit actor")
	ErrInvalidResource  = errors.New("invalid audit resource")
	ErrInvalidTimestamp = errors.New("invalid audit timestamp")
	ErrInvalidLimit     = errors.New("invalid audit limit")
	ErrInvalidCursor    = errors.New("invalid audit cursor")
	ErrDetailsTooLarge  = errors.New("audit details too large")
)

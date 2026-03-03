package audit

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidEventActorID   = errors.New("invalid audit event actor ID")
	ErrInvalidEventAction    = errors.New("invalid audit event action")
	ErrInvalidEventTimestamp = errors.New("invalid audit event timestamp")
	ErrInvalidEventDetails   = errors.New("invalid audit event details")
	ErrPayloadTooLarge       = errors.New("audit payload too large")
)

type change struct {
	OldValue any `json:"old_value,omitempty"`
	NewValue any `json:"new_value,omitempty"`
}

func processDetails(before, after map[string]any) map[string]any {
	changes := make(map[string]any)

	keys := make(map[string]struct{})
	for key := range before {
		keys[key] = struct{}{}
	}
	for key := range after {
		keys[key] = struct{}{}
	}

	for key := range keys {
		beforeValue, beforeExists := before[key]
		afterValue, afterExists := after[key]

		if !beforeExists {
			changes[key] = change{
				NewValue: afterValue,
			}
			continue
		}

		if !afterExists {
			changes[key] = change{
				OldValue: beforeValue,
			}
			continue
		}

		if beforeValue != afterValue {
			changes[key] = change{
				OldValue: beforeValue,
				NewValue: afterValue,
			}
		}
	}

	return changes
}

type event struct {
	ActorID   string         `json:"actorID" validate:"required,uuid4"`
	Timestamp time.Time      `json:"timestamp" validate:"required"`
	Action    string         `json:"action" validate:"required"`
	Details   map[string]any `json:"details" validate:"required"`
}

func validationErr(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			switch fieldErr.StructField() {
			case "ActorID":
				return ErrInvalidEventActorID
			case "Action":
				return ErrInvalidEventAction
			case "Timestamp":
				return ErrInvalidEventTimestamp
			case "Details":
				return ErrInvalidEventDetails
			default:
				return err
			}
		}
	}
	return err
}

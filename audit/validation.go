package audit

import (
	"errors"
	"fmt"
)

func ValidateEvent(event *Event, cfg *Configuration) error {
	if event == nil {
		return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrInvalidAction)
	}

	if cfg == nil {
		cfg = NewConfiguration()
	}

	if event.Action == "" {
		return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrInvalidAction)
	}

	if event.Actor.ID == "" || event.Actor.Type == "" {
		return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrInvalidActor)
	}

	if event.Resource.ID == "" || event.Resource.Type == "" {
		return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrInvalidResource)
	}

	if event.Timestamp.IsZero() {
		return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrInvalidTimestamp)
	}

	if _, err := MarshalBoundedJSON(event.Details, cfg.MaxDetailsBytes); err != nil {
		if errors.Is(err, ErrDetailsTooLarge) {
			return fmt.Errorf("%w: %w", ErrInvalidEvent, ErrDetailsTooLarge)
		}
		return fmt.Errorf("%w: %w", ErrInvalidEvent, err)
	}

	return nil
}

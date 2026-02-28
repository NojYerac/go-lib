package audit

import (
	"encoding/json"
	"fmt"
)

func MarshalBoundedJSON(payload any, maxBytes int) ([]byte, error) {
	if maxBytes < 0 {
		return nil, ErrInvalidLimit
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if maxBytes > 0 && len(data) > maxBytes {
		return nil, fmt.Errorf(
			"%w: payload size %d exceeds limit %d",
			ErrDetailsTooLarge,
			len(data),
			maxBytes,
		)
	}

	return data, nil
}

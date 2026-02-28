package auth

import "strings"

func BearerToken(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ErrMissingToken
	}

	parts := strings.Fields(trimmed)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidToken
	}
	if parts[1] == "" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Validator interface {
	Validate(context.Context, string) (*Claims, error)
}

type validator struct {
	config *Configuration
	nowFn  func() time.Time
}

type Option func(*validator)

func WithNow(nowFn func() time.Time) Option {
	return func(v *validator) {
		if nowFn != nil {
			v.nowFn = nowFn
		}
	}
}

func NewValidator(config *Configuration, opts ...Option) Validator {
	v := &validator{
		config: config,
		nowFn:  time.Now,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *validator) Validate(_ context.Context, token string) (*Claims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrMissingToken
	}
	if v == nil || v.config == nil {
		return nil, fmt.Errorf("auth validator not configured: %w", ErrInvalidToken)
	}
	if strings.TrimSpace(v.config.HMACSecret) == "" {
		return nil, fmt.Errorf("auth hmac secret is empty: %w", ErrInvalidToken)
	}

	parsedToken, err := jwt.Parse(
		token,
		func(parsedToken *jwt.Token) (any, error) {
			if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unsupported signing method %q: %w", parsedToken.Method.Alg(), ErrInvalidToken)
			}
			return []byte(v.config.HMACSecret), nil
		},
		jwt.WithAudience(v.config.Audience),
		jwt.WithIssuer(v.config.Issuer),
		jwt.WithLeeway(v.config.ClockSkew),
		jwt.WithTimeFunc(v.nowFn),
	)
	if err != nil {
		return nil, mapJWTError(err)
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claimsFromMap(mapClaims), nil
}

func claimsFromMap(claims jwt.MapClaims) *Claims {
	return &Claims{
		Subject:   claimString(claims, "sub"),
		Issuer:    claimString(claims, "iss"),
		Audience:  claimStringSlice(claims, "aud"),
		Roles:     claimStringSlice(claims, "roles"),
		TokenID:   claimString(claims, "jti"),
		ExpiresAt: claimTime(claims, "exp"),
		IssuedAt:  claimTime(claims, "iat"),
		NotBefore: claimTime(claims, "nbf"),
	}
}

func claimString(claims jwt.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func claimTime(claims jwt.MapClaims, key string) time.Time {
	value, ok := claims[key]
	if !ok || value == nil {
		return time.Time{}
	}
	seconds, ok := numericClaimToUnix(value)
	if !ok {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
}

func numericClaimToUnix(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func claimStringSlice(claims jwt.MapClaims, key string) []string {
	value, ok := claims[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return nil
		}
		return []string{typed}
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			s := fmt.Sprint(item)
			if s != "" {
				result = append(result, s)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	default:
		return nil
	}
}

func mapJWTError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, jwt.ErrTokenExpired) {
		return ErrTokenExpired
	}
	return fmt.Errorf("%w: %v", ErrInvalidToken, err)
}

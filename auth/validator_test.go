package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var (
		now    time.Time
		config *Configuration
	)

	BeforeEach(func() {
		now = time.Date(2026, time.February, 28, 12, 0, 0, 0, time.UTC)
		config = NewConfiguration()
		config.Issuer = "issuer.go-lib"
		config.Audience = "go-lib-auth"
		config.HMACSecret = "top-secret"
		config.ClockSkew = 5 * time.Second
	})

	It("validates a signed token and maps standard claims", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"sub":   "user-1",
			"iss":   config.Issuer,
			"aud":   []string{config.Audience},
			"roles": []string{"reader", "admin"},
			"jti":   "jwt-1",
			"iat":   now.Unix(),
			"nbf":   now.Add(-1 * time.Second).Unix(),
			"exp":   now.Add(2 * time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		claims, err := validator.Validate(context.Background(), token)

		Expect(err).NotTo(HaveOccurred())
		Expect(claims.Subject).To(Equal("user-1"))
		Expect(claims.Issuer).To(Equal(config.Issuer))
		Expect(claims.Audience).To(ContainElement(config.Audience))
		Expect(claims.Roles).To(ContainElements("reader", "admin"))
		Expect(claims.TokenID).To(Equal("jwt-1"))
		Expect(claims.IssuedAt.Unix()).To(Equal(now.Unix()))
		Expect(claims.ExpiresAt.Unix()).To(Equal(now.Add(2 * time.Minute).Unix()))
	})

	It("supports scalar audience and role claims", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"sub":   "user-2",
			"iss":   config.Issuer,
			"aud":   config.Audience,
			"roles": "writer",
			"exp":   now.Add(time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		claims, err := validator.Validate(context.Background(), token)

		Expect(err).NotTo(HaveOccurred())
		Expect(claims.Audience).To(Equal([]string{config.Audience}))
		Expect(claims.Roles).To(Equal([]string{"writer"}))
	})

	It("returns missing token when token is empty", func() {
		validator := NewValidator(config)

		_, err := validator.Validate(context.Background(), "")

		Expect(err).To(MatchError(ErrMissingToken))
	})

	It("returns invalid token when validator receiver is nil", func() {
		var v *validator
		_, err := v.Validate(context.Background(), "token")
		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("returns invalid token when config or secret is missing", func() {
		v := &validator{}
		_, err := v.Validate(context.Background(), "token")
		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))

		config.HMACSecret = ""
		validator := NewValidator(config)
		_, err = validator.Validate(context.Background(), "any")
		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("returns invalid token for wrong signature", func() {
		token := mustSignHMACToken("other-secret", jwt.MapClaims{
			"iss": config.Issuer,
			"aud": config.Audience,
			"exp": now.Add(time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		_, err := validator.Validate(context.Background(), token)

		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("returns invalid token for wrong issuer", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"iss": "other-issuer",
			"aud": config.Audience,
			"exp": now.Add(time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		_, err := validator.Validate(context.Background(), token)

		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("returns invalid token for wrong audience", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"iss": config.Issuer,
			"aud": "different-audience",
			"exp": now.Add(time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		_, err := validator.Validate(context.Background(), token)

		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("returns token expired when token is outside skew window", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"iss": config.Issuer,
			"aud": config.Audience,
			"exp": now.Add(-config.ClockSkew - time.Second).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		_, err := validator.Validate(context.Background(), token)

		Expect(err).To(MatchError(ErrTokenExpired))
	})

	It("accepts a token inside clock skew", func() {
		token := mustSignHMACToken(config.HMACSecret, jwt.MapClaims{
			"sub": "user-3",
			"iss": config.Issuer,
			"aud": config.Audience,
			"exp": now.Add(-config.ClockSkew + time.Second).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		claims, err := validator.Validate(context.Background(), token)

		Expect(err).NotTo(HaveOccurred())
		Expect(claims.Subject).To(Equal("user-3"))
	})

	It("returns invalid token for unsupported signing method", func() {
		token := mustSignNoneToken(jwt.MapClaims{
			"iss": config.Issuer,
			"aud": config.Audience,
			"exp": now.Add(time.Minute).Unix(),
		})
		validator := NewValidator(config, WithNow(func() time.Time { return now }))

		_, err := validator.Validate(context.Background(), token)

		Expect(err).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})

	It("covers helper branches for claim conversion and error mapping", func() {
		WithNow(nil)(&validator{nowFn: func() time.Time { return now }})

		claims := claimsFromMap(jwt.MapClaims{
			"sub":   "user-4",
			"iss":   "issuer-4",
			"aud":   []any{"svc-a", "svc-b"},
			"roles": []any{"admin", 1},
			"jti":   "jti-4",
			"exp":   json.Number("500"),
			"iat":   "501",
			"nbf":   502,
		})
		Expect(claims.Subject).To(Equal("user-4"))
		Expect(claims.Audience).To(Equal([]string{"svc-a", "svc-b"}))
		Expect(claims.Roles).To(Equal([]string{"admin", "1"}))
		Expect(claims.ExpiresAt.Unix()).To(Equal(int64(500)))

		n, ok := numericClaimToUnix(struct{}{})
		Expect(ok).To(BeFalse())
		Expect(n).To(BeZero())

		Expect(claimTime(jwt.MapClaims{"exp": struct{}{}}, "exp").IsZero()).To(BeTrue())
		Expect(claimStringSlice(jwt.MapClaims{"roles": ""}, "roles")).To(BeNil())
		Expect(claimStringSlice(jwt.MapClaims{"roles": 1}, "roles")).To(BeNil())

		Expect(mapJWTError(nil)).To(Succeed())
		Expect(mapJWTError(jwt.ErrTokenExpired)).To(MatchError(ErrTokenExpired))
		Expect(mapJWTError(errors.New("boom"))).To(MatchError(ContainSubstring(ErrInvalidToken.Error())))
	})
})

func mustSignHMACToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signedToken
}

func mustSignNoneToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		panic(err)
	}
	return signedToken
}

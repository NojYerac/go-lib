package testtoken

import (
	"time"

	gojwt "github.com/dgrijalva/jwt-go"
	"github.com/nojyerac/go-lib/pkg/auth"
	"github.com/nojyerac/go-lib/pkg/jwt"
)

//nolint:gosec // key not to be used outside of testing
const (
	PrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIA4WF79lBYQCjjIOunx5N75WdqstUwI4XYIqLZSxyJtqoAoGCCqGSM49
AwEHoUQDQgAEyViEkF0WSOVwYcISC9bokDxVVibYftwFC/YY3Q3oXDX0iAD3waIm
4J9yN4gD1K8kmQN80jfjSjf2k2hLhZ1X3Q==
-----END EC PRIVATE KEY-----`

	PublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEyViEkF0WSOVwYcISC9bokDxVVibY
ftwFC/YY3Q3oXDX0iAD3waIm4J9yN4gD1K8kmQN80jfjSjf2k2hLhZ1X3Q==
-----END PUBLIC KEY-----`
)

var (
	Issuer   jwt.Issuer
	Verifier jwt.Verifier
)

func init() {
	Issuer = jwt.NewIssuer(&jwt.Configuration{
		AccessTokenTTL:   time.Minute,
		JWTSigningMethod: gojwt.SigningMethodES256.Name,
		JWTPrivateKey:    PrivateKey,
	})

	Verifier = jwt.NewVerifier(&jwt.Configuration{
		AccessTokenTTL:   time.Minute,
		JWTSigningMethod: gojwt.SigningMethodES256.Name,
		JWTPublicKey:     PublicKey,
	})
}

func TestToken() string {
	token, _ := Issuer.AccessToken(&auth.User{
		UserID:    99,
		Username:  "testuser",
		Privleges: []string{"TEST_PRIV"},
		Features:  []string{"TEST_FEATURE"},
	})
	return token
}

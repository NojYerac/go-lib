# JWT Package

The **jwt** package offers helpers for generating and validating JSON Web Tokens. It supports ES series signing methods and allows configuration of token TTL and keys.

## Configuration

```go
// pkg/jwt/config.go
package jwt

import "time"

// Configuration holds the JWT settings.
//
//   AccessTokenTTL   time.Duration `config:"access_token_ttl" validate:"required"`
//   JWTSigningMethod string        `config:"jwt_signing_method" validate:"required,oneof=ES256 ES384 ES512"`
//   JWTPrivateKey    string        `config:"jwt_private_key" validate:"required_without=JWTPublicKey,priv_ec_key"`
//   JWTPublicKey     string        `config:"jwt_public_key" validate:"required_without=JWTPrivateKey,pub_key"`
//
// The NewConfiguration helper returns sane defaults.
```

### Token Generation

```go
func NewToken(claims jwt.Claims, cfg *jwt.Configuration) (string, error) {
    // Load private key, sign the token, and return the signed string.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/jwt"
    "github.com/dgrijalva/jwt-go"
)

func main() {
    cfg := jwt.NewConfiguration()
    // Load keys into cfg.JWTPrivateKey / cfg.JWTPublicKey
    claims := jwt.MapClaims{"sub": "1234", "exp": time.Now().Add(cfg.AccessTokenTTL).Unix()}
    token, err := jwt.NewToken(claims, cfg)
    if err != nil {
        // handle error
    }
    println("access token:", token)
}
```

## Examples

- Create a token with a custom payload.
- Validate a token against the configured public key.
- Refresh tokens with a new TTL.

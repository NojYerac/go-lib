package token

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Configuration struct {
	AccessTokenTTL   time.Duration `config:"access_token_ttl" validate:"required"`
	RefreshTokenTTL  time.Duration `config:"refresh_token_ttl"`
	JWTSigningMethod string        `config:"jwt_signing_method" validate:"required,oneof=ES256 ES384 ES512"`
	JWTPrivateKey    string        `config:"jwt_private_key" validate:"required,priv_ec_key"`
	JWTPublicKey     string        `config:"jwt_public_key" validate:"required,pub_key"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		JWTSigningMethod: jwt.SigningMethodES256.Name,
	}
}

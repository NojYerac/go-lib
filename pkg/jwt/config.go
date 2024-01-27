package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Configuration struct {
	AccessTokenTTL   time.Duration `config:"access_token_ttl" validate:"required"`
	JWTSigningMethod string        `config:"jwt_signing_method" validate:"required,oneof=ES256 ES384 ES512"`
	JWTPrivateKey    string        `config:"jwt_private_key" validate:"required_without=JWTPublicKey,priv_ec_key"`
	JWTPublicKey     string        `config:"jwt_public_key" validate:"required_without=JWTPrivateKey,pub_key"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		AccessTokenTTL:   15 * time.Minute,
		JWTSigningMethod: jwt.SigningMethodES256.Name,
	}
}

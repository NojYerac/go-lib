package token

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"source.rad.af/libs/go-lib/pkg/auth"
	"source.rad.af/libs/go-lib/pkg/version"
)

var (
	errUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// Session is the data struct for a session
type Session struct {
	auth.User
	jwt.StandardClaims
}

func (s *Session) Valid() error {
	return s.StandardClaims.Valid()
}

type Issuer interface {
	AccessToken(s *auth.User) (string, error)
	Session(string) (*Session, error)
}

func NewIssuer(config *Configuration) Issuer {
	var signingMethod jwt.SigningMethod
	switch config.JWTSigningMethod {
	case jwt.SigningMethodES256.Name:
		signingMethod = jwt.SigningMethodES256
	case jwt.SigningMethodES384.Name:
		signingMethod = jwt.SigningMethodES384
	case jwt.SigningMethodES512.Name:
		signingMethod = jwt.SigningMethodES512
	default:
		signingMethod = jwt.SigningMethodNone
	}
	return &issuer{
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
		privateKey:      config.JWTPrivateKey,
		publicKey:       config.JWTPublicKey,
		signingMethod:   signingMethod,
	}
}

type issuer struct {
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	signingMethod   jwt.SigningMethod
	privateKey      string
	publicKey       string
}

// AccessToken returns a signed JWT
func (i *issuer) AccessToken(user *auth.User) (string, error) {
	s := &Session{
		User: *user,
	}
	now := jwt.TimeFunc()
	s.Issuer = version.GetVersion().Name
	s.NotBefore = now.Unix() - 60
	s.IssuedAt = now.Unix() - 60
	s.ExpiresAt = now.Add(i.accessTokenTTL).Unix()
	token := jwt.NewWithClaims(i.signingMethod, s)
	return token.SignedString(i.privateKey)
}

// Session parses a signed JWT to a session
func (i *issuer) Session(signed string) (*Session, error) {
	token, err := jwt.ParseWithClaims(signed, &Session{}, i.keyFunc)
	if err != nil {
		return nil, err
	}
	session := token.Claims.(*Session)
	return session, nil
}

func (i *issuer) keyFunc(t *jwt.Token) (interface{}, error) {
	if t.Method != i.signingMethod {
		return nil, errUnexpectedSigningMethod
	}
	return i.publicKey, nil
}

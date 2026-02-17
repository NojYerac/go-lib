package jwt

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nojyerac/go-lib/pkg/auth"
	"github.com/nojyerac/go-lib/pkg/version"
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
	Verifier
	AccessToken(s *auth.User) (string, error)
}

type Verifier interface {
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
	var privateKey *ecdsa.PrivateKey
	var publicKey *ecdsa.PublicKey
	var err error
	if len(config.JWTPrivateKey) > 0 {
		privBlock, _ := pem.Decode([]byte(config.JWTPrivateKey))
		privateKey, err = x509.ParseECPrivateKey(privBlock.Bytes)
		if err != nil {
			panic(err)
		}
	}
	if len(config.JWTPublicKey) > 0 {
		pubBlock, _ := pem.Decode([]byte(config.JWTPublicKey))
		genericPub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		if err != nil {
			panic(err)
		}
		publicKey = genericPub.(*ecdsa.PublicKey)
	} else {
		publicKey = &privateKey.PublicKey
	}

	return &issuer{
		accessTokenTTL: config.AccessTokenTTL,
		privateKey:     privateKey,
		publicKey:      publicKey,
		signingMethod:  signingMethod,
	}
}

func NewVerifier(config *Configuration) Verifier {
	return NewIssuer(config)
}

type issuer struct {
	accessTokenTTL time.Duration
	signingMethod  jwt.SigningMethod
	privateKey     *ecdsa.PrivateKey
	publicKey      *ecdsa.PublicKey
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

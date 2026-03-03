package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var customValidators = map[string]validator.Func{
	"pub_key": func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return strings.HasPrefix(s, "-----BEGIN PUBLIC KEY-----") &&
			strings.HasSuffix(s, "-----END PUBLIC KEY-----")
	},
	"priv_ec_key": func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return strings.HasPrefix(s, "-----BEGIN EC PRIVATE KEY-----") &&
			strings.HasSuffix(s, "-----END EC PRIVATE KEY-----")
	},
}

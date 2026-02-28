package auth

import "time"

type Claims struct {
	Subject   string
	Issuer    string
	Audience  []string
	Roles     []string
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
	TokenID   string
}

func (c *Claims) HasRole(role string) bool {
	if c == nil {
		return false
	}
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *Claims) HasAnyRole(roles ...string) bool {
	if c == nil {
		return false
	}
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

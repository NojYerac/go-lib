package authz

import "github.com/nojyerac/go-lib/auth"

type Requirement struct {
	AnyOf []string
	AllOf []string
}

func RequireAny(roles ...string) Requirement {
	return Requirement{AnyOf: normalizeRoles(roles)}
}

func RequireAll(roles ...string) Requirement {
	return Requirement{AllOf: normalizeRoles(roles)}
}

func (r Requirement) IsEmpty() bool {
	return len(r.AnyOf) == 0 && len(r.AllOf) == 0
}

func (r Requirement) SatisfiedBy(claims *auth.Claims) bool {
	if r.IsEmpty() {
		return true
	}
	if claims == nil {
		return false
	}

	if len(r.AllOf) > 0 {
		for _, requiredRole := range r.AllOf {
			if !claims.HasRole(requiredRole) {
				return false
			}
		}
	}

	if len(r.AnyOf) == 0 {
		return true
	}

	return claims.HasAnyRole(r.AnyOf...)
}

func Authorize(claims *auth.Claims, requirement Requirement) error {
	if requirement.SatisfiedBy(claims) {
		return nil
	}
	return auth.ErrPermissionDenied
}

func normalizeRoles(roles []string) []string {
	if len(roles) == 0 {
		return nil
	}

	result := make([]string, 0, len(roles))
	seen := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

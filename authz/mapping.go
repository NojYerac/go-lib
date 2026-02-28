package authz

import (
	"fmt"
	"strings"
)

type PolicyMap map[string]Requirement

func NewPolicyMap() PolicyMap {
	return make(PolicyMap)
}

func (p PolicyMap) Set(operation string, requirement Requirement) {
	if p == nil {
		return
	}
	key := normalizeOperationKey(operation)
	if key == "" {
		return
	}
	p[key] = requirement
}

func (p PolicyMap) Requirement(operation string) (Requirement, bool) {
	if p == nil {
		return Requirement{}, false
	}
	requirement, ok := p[normalizeOperationKey(operation)]
	return requirement, ok
}

func HTTPOperation(method, path string) string {
	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedPath := strings.TrimSpace(path)
	if normalizedMethod == "" || normalizedPath == "" {
		return ""
	}
	return fmt.Sprintf("%s %s", normalizedMethod, normalizedPath)
}

func GRPCOperation(fullMethod string) string {
	return normalizeOperationKey(fullMethod)
}

func normalizeOperationKey(operation string) string {
	return strings.TrimSpace(operation)
}

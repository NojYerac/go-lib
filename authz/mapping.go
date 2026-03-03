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
	normalizedOperation := normalizeOperationKey(operation)
	requirement, ok := p[normalizedOperation]
	if ok {
		return requirement, true
	}

	method, path, isHTTP := parseHTTPOperation(normalizedOperation)
	if !isHTTP {
		return Requirement{}, false
	}

	bestScore := -1
	bestRequirement := Requirement{}
	for candidateOperation, candidateRequirement := range p {
		candidateMethod, candidatePath, candidateIsHTTP := parseHTTPOperation(candidateOperation)
		if !candidateIsHTTP || candidateMethod != method {
			continue
		}

		score, matches := matchHTTPPath(path, candidatePath)
		if !matches || score <= bestScore {
			continue
		}

		bestScore = score
		bestRequirement = candidateRequirement
	}

	if bestScore < 0 {
		return Requirement{}, false
	}

	return bestRequirement, true
}

func HTTPOperation(method, path string) string {
	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedPath := normalizeHTTPPath(path)
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

func parseHTTPOperation(operation string) (method, path string, ok bool) {
	methodPart, pathPart, found := strings.Cut(strings.TrimSpace(operation), " ")
	if !found {
		return "", "", false
	}

	method = strings.ToUpper(strings.TrimSpace(methodPart))
	path = normalizeHTTPPath(pathPart)
	if method == "" || path == "" {
		return "", "", false
	}

	return method, path, true
}

func normalizeHTTPPath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return ""
	}

	if queryOrFragmentAt := strings.IndexAny(normalized, "?#"); queryOrFragmentAt >= 0 {
		normalized = normalized[:queryOrFragmentAt]
	}

	if normalized == "" {
		return ""
	}

	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	if normalized != "/" {
		normalized = strings.TrimRight(normalized, "/")
	}

	return normalized
}

func matchHTTPPath(requestPath, policyPath string) (int, bool) {
	requestSegments := splitPathSegments(requestPath)
	policySegments := splitPathSegments(policyPath)
	if len(requestSegments) != len(policySegments) {
		return 0, false
	}

	score := 0
	for i := range requestSegments {
		if isPathParamSegment(policySegments[i]) {
			continue
		}
		if requestSegments[i] != policySegments[i] {
			return 0, false
		}
		score++
	}

	return score, true
}

func splitPathSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func isPathParamSegment(segment string) bool {
	if segment == "*" {
		return true
	}

	if strings.HasPrefix(segment, ":") && len(segment) > 1 {
		return true
	}

	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2
}

package route

import (
	"fmt"
	"regexp"
	"strings"
)

var segmentPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// NormalizeSegment validates a relative group segment (no slashes).
// An empty segment is allowed for middleware-only groups.
func NormalizeSegment(segment string) (string, error) {
	if strings.Contains(segment, "/") {
		return "", fmt.Errorf("route: segment %q must not contain slashes", segment)
	}
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return "", nil
	}
	if !segmentPattern.MatchString(segment) {
		return "", fmt.Errorf("route: invalid segment %q: must match [a-z][a-z0-9_-]*", segment)
	}
	return segment, nil
}

// NormalizePattern validates and normalizes a route pattern relative to the current group.
// Literal pages and collection paths end with /. Patterns ending in {$} are unchanged.
func NormalizePattern(pattern string) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "/" {
		return "/", nil
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	if err := validatePatternSegments(pattern); err != nil {
		return "", err
	}

	if strings.HasSuffix(pattern, "{$}") {
		return pattern, nil
	}

	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	return pattern, nil
}

func validatePatternSegments(pattern string) error {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if part == "{$}" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			if name != "id" && name != "edge" {
				return fmt.Errorf("route: unsupported path parameter %q: only {id} and {edge} are allowed", part)
			}
			continue
		}
		if !segmentPattern.MatchString(part) {
			return fmt.Errorf("route: invalid path segment %q: must match [a-z][a-z0-9_-]*", part)
		}
	}
	return nil
}

// NormalizeMountPrefix validates an absolute mount prefix such as /admin/.
func NormalizeMountPrefix(prefix string) (mount string, strip string, err error) {
	if prefix == "" {
		return "", "", fmt.Errorf("route: mount prefix is required")
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	mount = prefix
	if !strings.HasSuffix(mount, "/") {
		mount += "/"
	}
	strip = strings.TrimSuffix(mount, "/")
	if strip == "" {
		return "", "", fmt.Errorf("route: mount prefix must not be /")
	}
	return mount, strip, nil
}

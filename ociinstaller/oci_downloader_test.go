package ociinstaller

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"oras.land/oras-go/v2/registry/remote/errcode"
)

func TestRegistryHostFromRef(t *testing.T) {
	cases := map[string]string{
		"ghcr.io/turbot/steampipe/plugins/turbot/aws:1.30.2":  "ghcr.io",
		"ghcr.io/turbot/steampipe/plugins/turbot/aws:latest": "ghcr.io",
		"hub.steampipe.io/plugins/turbot/aws:latest":         "hub.steampipe.io",
		"registry.example.com:5000/org/image:tag":            "registry.example.com:5000",
		"ghcr.io":                                            "ghcr.io",
		"":                                                   "",
	}
	for ref, want := range cases {
		if got := registryHostFromRef(ref); got != want {
			t.Errorf("registryHostFromRef(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestIsRegistryAuthError(t *testing.T) {
	mkErr := func(status int) error {
		u, _ := url.Parse("https://ghcr.io/token")
		return &errcode.ErrorResponse{
			Method:     "GET",
			URL:        u,
			StatusCode: status,
			Errors:     errcode.Errors{{Code: errcode.ErrorCodeDenied, Message: "denied"}},
		}
	}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("network unreachable"), false},
		{"401 Unauthorized", mkErr(http.StatusUnauthorized), true},
		{"403 Forbidden", mkErr(http.StatusForbidden), true},
		{"404 Not Found", mkErr(http.StatusNotFound), false},
		{"500 Server Error", mkErr(http.StatusInternalServerError), false},
		{"wrapped 403", fmt.Errorf("resolving manifest: %w", mkErr(http.StatusForbidden)), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRegistryAuthError(tc.err); got != tc.want {
				t.Errorf("isRegistryAuthError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestWrapAuthError(t *testing.T) {
	inner := errors.New("response status code 403: denied: denied")
	wrapped := wrapAuthError(inner, "ghcr.io/turbot/steampipe/plugins/turbot/aws:1.30.2")

	msg := wrapped.Error()
	if !contains(msg, "denied: denied") {
		t.Errorf("wrapped error lost original message, got: %s", msg)
	}
	if !contains(msg, "ghcr.io") {
		t.Errorf("wrapped error missing host hint, got: %s", msg)
	}
	if !contains(msg, "docker logout") {
		t.Errorf("wrapped error missing docker logout hint, got: %s", msg)
	}
	if !errors.Is(wrapped, inner) {
		t.Errorf("errors.Is should still see the original error through the wrap")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

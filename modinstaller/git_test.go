package modinstaller

import (
	"testing"
)

func TestTransformToGitURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		urlMode  GitUrlMode
		expected string
	}{
		// GitHub - SSH mode
		{"GitHub SSH - Basic", "github.com/user/repo", GitUrlModeSSH, "git@github.com:user/repo.git"},
		{"GitHub SSH - With Git Prefix", "git@github.com/user/repo", GitUrlModeSSH, "git@github.com:user/repo.git"},
		{"GitHub SSH - With .git", "github.com/user/repo.git", GitUrlModeSSH, "git@github.com:user/repo.git"},

		// GitHub - HTTPS mode
		{"GitHub HTTPS - Basic", "github.com/user/repo", GitUrlModeHTTPS, "https://github.com/user/repo"},
		{"GitHub HTTPS - Already HTTPS", "https://github.com/user/repo", GitUrlModeHTTPS, "https://github.com/user/repo"},

		// GitLab - SSH mode
		{"GitLab SSH - Basic", "gitlab.com/user/repo", GitUrlModeSSH, "git@gitlab.com:user/repo.git"},
		{"GitLab SSH - With Git Prefix", "git@gitlab.com/user/repo", GitUrlModeSSH, "git@gitlab.com:user/repo.git"},
		{"GitLab SSH - With .git", "gitlab.com/user/repo.git", GitUrlModeSSH, "git@gitlab.com:user/repo.git"},

		// GitLab - HTTPS mode
		{"GitLab HTTPS - Basic", "gitlab.com/user/repo", GitUrlModeHTTPS, "https://gitlab.com/user/repo"},
		{"GitLab HTTPS - Already HTTPS", "https://gitlab.com/user/repo", GitUrlModeHTTPS, "https://gitlab.com/user/repo"},
		{"GitLab HTTPS - with domain name", "example.gitlab.com/user/repo", GitUrlModeHTTPS, "https://example.gitlab.com/user/repo"},

		// Edge cases
		{"Unsupported Host", "other-host.com/user/repo", GitUrlModeSSH, "other-host.com/user/repo"},
		{"Invalid Input - Empty", "", GitUrlModeSSH, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformToGitURL(tt.input, tt.urlMode)
			if got != tt.expected {
				t.Errorf("transformToGitURL(%q, %q) = %q; want %q", tt.input, tt.urlMode, got, tt.expected)
			}
		})
	}
}

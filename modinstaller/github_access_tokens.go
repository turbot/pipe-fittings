package modinstaller

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"strings"
)

const (
PersonalAccessTokenPrefix="ghp_"
FineGrainedPersonalAccessTokenPrefix="github_pat_"
OAuthAccessTokenPrefix	= "gho_"
GitHubAppUserAccessTokenPrefix = "ghu_"
GitHubAppInstallationAccessTokenPrefix = "ghs_"
GitHubAppRefreshTokenPrefix = "ghr_"
)

func GetAuthForGithubToken(token string) transport.AuthMethod {
	if token == "" {
		return nil
	}
	SWAP - check for app specifically and if so use go-github????
	// personal tokens use basic auth
	if strings.HasPrefix(token, GitHubAppUserAccessTokenPrefix) ||
		strings.HasPrefix(token, GitHubAppInstallationAccessTokenPrefix) ||
		strings.HasPrefix(token, GitHubAppRefreshTokenPrefix) {
		return &http.BasicAuth{
			Username: token,
		}
	}

	// app tokens require bearer token
	return &http.TokenAuth{
		Token:  token,
	}
}
package modinstaller

import (
	"context"
	"github.com/google/go-github/v61/github"
	"net/http"
	"strings"
)

func getTagsUsingGithubAPI(modPath, token string) ([]string, error) {
	// Create client to access the GitHub API
	// If a token is passed, use it to create the client, otherwise use the default client
	var client *github.Client
	if token != "" {
		if strings.HasPrefix(token, GitHubAppInstallationAccessTokenPrefix) {
			client = github.NewClient(&http.Client{Transport: &oauth2Transport{
				Token: token,
			}})
		} else {
			client = github.NewClient(nil).WithAuthToken(token)
		}
	} else {
		client = github.NewClient(nil)
	}

	owner, name, err := getOwnerAndOrgFromGitUrl(modPath)
	if err != nil {
		return nil, err
	}
	// load remote references
	refs, _, err := client.Git.ListMatchingRefs(context.Background(), owner, name, nil)
	if err != nil {
		return nil, err
	}

	// filters the references list and only keeps tags
	var tags []string
	for _, ref := range refs {
		if ref.Object.GetType() == "tag" {
			refName := ref.GetRef()
			tags = append(tags, getShortRefName(refName))
		}
	}

	return tags, nil
}

// oauth2Transport is an http.RoundTripper that authenticates all requests
type oauth2Transport struct {
	Token string
}

func (t *oauth2Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.Token)
	return http.DefaultTransport.RoundTrip(clone)
}

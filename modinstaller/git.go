package modinstaller

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/turbot/pipe-fittings/app_specific"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitUrlMode int

const (
	GitUrlModeHTTPS GitUrlMode = iota
	GitUrlModeSSH
)

func getGitUrl(modName string, urlMode GitUrlMode) string {
	return transformToGitURL(modName, urlMode)
}

func transformToGitURL(input string, urlMode GitUrlMode) string {
	if urlMode == GitUrlModeHTTPS {
		if !strings.HasPrefix(input, "https://") {
			input = "https://" + input
		}
		return input
	}

	if !strings.HasPrefix(input, "github.com") {
		return input
	}

	if !strings.HasPrefix(input, "git@") {
		input = "git@" + input
	}

	if !strings.HasSuffix(input, ".git") {
		input += ".git"
	}

	// Add a colon after the "git@github.com" part, so it replaces the first / with :
	if !strings.Contains(input, ":") {
		index := strings.Index(input, "/")
		input = input[:index] + ":" + input[index+1:]
	}

	return input
}

func getTags(repo string) ([]string, error) {
	gitHubToken := getGitToken()

	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	var listOption git.ListOptions
	// if a token was provided, use it
	if gitHubToken != "" {
		listOption = git.ListOptions{
			Auth: getGitAuthForToken(gitHubToken),
		}
	}
	// load remote references
	refs, err := rem.List(&listOption)
	if err != nil {
		return nil, err
	}

	// filters the references list and only keeps tags
	var tags []string
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tags = append(tags, ref.Name().Short())
		}
	}

	return tags, nil
}

func getGitAuthForToken(gitHubToken string) transport.AuthMethod {
	if gitHubToken == "" {
		return nil
	}
	var auth transport.AuthMethod
	// if authentication token is an app token, we need to use the GitHub API to list
	if strings.HasPrefix(gitHubToken, GitHubAppInstallationAccessTokenPrefix) {
		// (NOTE: set user to x-access-token - this is required for github application tokens))
		auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: gitHubToken,
		}
	} else {
		auth = &http.BasicAuth{
			Username: gitHubToken,
		}
	}
	return auth
}

func getGitToken() string {
	return os.Getenv(app_specific.EnvGitToken)
}

func getTagVersionsFromGit(modName string, includePrerelease bool) (semver.Collection, error) {
	slog.Debug("getTagVersionsFromGit - retrieving tags from Git", "mod", modName)
	// first try https
	repo := getGitUrl(modName, GitUrlModeHTTPS)

	slog.Debug("trying HTTPS", "url", repo)
	tags, err := getTags(repo)
	if err != nil {
		slog.Debug("HTTPS failed", "error", err)
		// if that fails try ssh
		repo = getGitUrl(modName, GitUrlModeSSH)

		slog.Debug("trying SSH", "url", repo)
		tags, err = getTags(repo)
		if err != nil {
			slog.Debug("SSH failed", "error", err)
			return nil, err
		}
	}

	slog.Debug("retrieved tags from Git")

	versions := make(semver.Collection, len(tags))
	// handle index manually as we may not add all tags - if we cannot parse them as a version
	idx := 0
	for _, raw := range tags {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}

		if (!includePrerelease && v.Metadata() != "") || (!includePrerelease && v.Prerelease() != "") {
			continue
		}
		versions[idx] = v
		idx++
	}
	// shrink slice
	versions = versions[:idx]

	// sort the versions in REVERSE order
	sort.Sort(sort.Reverse(versions))
	return versions, nil
}

package modinstaller

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (i *ModInstaller[T]) openRepo(modPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(modPath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (i *ModInstaller[T]) cloneRepo(gitUrl string, gitRefName plumbing.ReferenceName, installPath string) (*git.Repository, error) {
	gitHubToken := getGitToken()
	cloneOptions := git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: gitRefName,
		Depth:         1,
		SingleBranch:  true,
	}
	if gitHubToken != "" {
		// if we have a token, use it
		cloneOptions.Auth = getGitAuthForToken(gitHubToken)
	}

	repo, err := git.PlainClone(installPath,
		false, &cloneOptions)
	return repo, err
}

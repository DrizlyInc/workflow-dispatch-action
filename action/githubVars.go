package main

import (
	"errors"
	"os"
	"strings"
)

// https://docs.github.com/en/actions/reference/environment-variables#default-environment-variables
type githubVars struct {
	repository      string
	repositoryOwner string
	repositoryName  string
	sha             string
	apiUrl          string
	serverUrl       string
}

func parseGithubVars() (githubVars, error) {

	repository, ok := os.LookupEnv("GITHUB_REPOSITORY")
	if !ok {
		return githubVars{}, errors.New("GITHUB_REPOSITORY env var not set")
	}
	ownerRepoSplit := strings.Split(repository, "/")
	if len(ownerRepoSplit) != 2 {
		return githubVars{}, errors.New("GITHUB_REPOSITORY env var not formatted as owner/repo-name")
	}
	repositoryOwner := ownerRepoSplit[0]
	repositoryName := ownerRepoSplit[1]

	sha, ok := os.LookupEnv("GITHUB_SHA")
	if !ok {
		return githubVars{}, errors.New("GITHUB_SHA env var not set")
	}

	apiUrl, ok := os.LookupEnv("GITHUB_API_URL")
	if !ok {
		return githubVars{}, errors.New("GITHUB_API_URL env var not set")
	}

	serverUrl, ok := os.LookupEnv("GITHUB_SERVER_URL")
	if !ok {
		return githubVars{}, errors.New("GITHUB_SERVER_URL env var not set")
	}

	return githubVars{
		repository:      repository,
		repositoryOwner: repositoryOwner,
		repositoryName:  repositoryName,
		sha:             sha,
		apiUrl:          apiUrl,
		serverUrl:       serverUrl,
	}, nil
}

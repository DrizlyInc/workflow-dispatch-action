package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v37/github"
)

func newGithubClient(tr http.RoundTripper, githubVars githubVars, inputs inputs) (*github.Client, error) {

	// https://github.com/google/go-github#authentication
	// First, create an AppsTransport for initial auth
	itr := ghinstallation.NewAppsTransportFromPrivateKey(tr, inputs.appID, inputs.privateKey)
	itr.BaseURL = githubVars.apiUrl

	// use appTransport to generate a client
	client := github.NewClient(&http.Client{Transport: itr})

	// We only need 1 installation since the app we are interested in only exists
	// in the context of its 1 installation to our org/repo
	opt := &github.ListOptions{
		PerPage: 1,
	}
	installations, _, err := client.Apps.ListInstallations(context.Background(), opt)
	if err != nil {
		return nil, fmt.Errorf("Error fetching app installations: %w", err)
	}

	// Take the first (only) installation we fetched
	ntr := ghinstallation.NewFromAppsTransport(itr, *installations[0].ID)
	return github.NewClient(&http.Client{Transport: ntr}), nil
}

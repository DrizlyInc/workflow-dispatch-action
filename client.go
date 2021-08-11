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

	// Get the list of installations
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var allInstallations []*github.Installation
	for {
		installations, resp, err := client.Apps.ListInstallations(context.Background(), opt)
		if err != nil {
			return nil, fmt.Errorf("error getting installations: %w", err)
		}
		allInstallations = append(allInstallations, installations...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// search for the specific installation we care about
	// spew.Dump(allInstallations)

	// construct client with the installation
	ntr := ghinstallation.NewFromAppsTransport(itr, *allInstallations[0].ID)
	return github.NewClient(&http.Client{Transport: ntr}), nil
}

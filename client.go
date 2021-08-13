package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
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

func fetchCheckWithRetries(ctx context.Context, client *github.Client, githubVars githubVars, checkId int) (*github.CheckRun, error) {
	secondsBetweenAttempts := 1
	maxAttempts := 4
	attempts := 0

	for {
		check, _, err := client.Checks.GetCheckRun(ctx, githubVars.repositoryOwner, githubVars.repositoryName, int64(checkId))
		attempts += 1

		if err != nil {
			githubactions.Warningf("Error fetching check %v (attempt %v of %v): %w", checkId, attempts, maxAttempts, err)
			time.Sleep(time.Second * time.Duration(secondsBetweenAttempts))
		} else {
			return check, nil
		}

		if attempts == maxAttempts {
			return nil, fmt.Errorf("Exceeded max attempts fetching check %v: %w", checkId, err)
		}
	}

}

func validateTargetWorkflowExistsOnDefaultBranch(ctx context.Context, client *github.Client, githubVars githubVars, inputs inputs) {
	targetRepository, _, err := client.Repositories.Get(ctx, inputs.targetOwner, inputs.targetRepository)
	if err != nil {
		githubactions.Fatalf("Failed to fetch target repository information: %w", err)
	}

	targetWorkflowFilepath := fmt.Sprintf(".github/workflows/%v.yml", inputs.workflowFilename)
	_, _, _, err = client.Repositories.GetContents(ctx, inputs.targetOwner, inputs.targetRepository, targetWorkflowFilepath, &github.RepositoryContentGetOptions{
		Ref: *targetRepository.DefaultBranch,
	})
	if err != nil {
		// https://github.com/DrizlyInc/distillery/blob/main/.github/workflows/tutorial00.yml
		expectedFileLocation := fmt.Sprintf("%v/%v/%v/blob/%v/%v", githubVars.serverUrl, inputs.targetOwner, inputs.targetRepository, *targetRepository.DefaultBranch, targetWorkflowFilepath)
		githubactions.Fatalf("The target workflow must exist on the default branch of the target repository!\nExpected to find it at: %v", expectedFileLocation)
	}
}

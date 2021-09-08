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
			githubactions.Warningf("Error fetching check %v (attempt %v of %v): %v", checkId, attempts, maxAttempts, err.Error())
			time.Sleep(time.Second * time.Duration(secondsBetweenAttempts))
		} else {
			return check, nil
		}

		if attempts == maxAttempts {
			return nil, fmt.Errorf("Exceeded max attempts fetching check %v: %w", checkId, err)
		}
	}

}

func validateTargetWorkflowExists(ctx context.Context, client *github.Client, githubVars githubVars, inputs inputs) {
	targetRepository, _, fetchRepoErr := client.Repositories.Get(ctx, inputs.targetOwner, inputs.targetRepository)
	if fetchRepoErr != nil {
		githubactions.Fatalf("Failed to fetch target repository information: %v", fetchRepoErr.Error())
	}

	workflowFilepath := fmt.Sprintf(".github/workflows/%v.yml", inputs.workflowFilename)
	_, _, _, fetchDefaultBranchWorkflowErr := client.Repositories.GetContents(ctx, inputs.targetOwner, inputs.targetRepository, workflowFilepath, &github.RepositoryContentGetOptions{
		Ref: *targetRepository.DefaultBranch,
	})
	if fetchDefaultBranchWorkflowErr != nil {
		// https://github.com/DrizlyInc/distillery/blob/main/.github/workflows/tutorial00.yml
		expectedFileUrl := fmt.Sprintf("%v/%v/%v/blob/%v/%v", githubVars.serverUrl, inputs.targetOwner, inputs.targetRepository, *targetRepository.DefaultBranch, workflowFilepath)
		githubactions.Errorf("The target workflow must exist on the default branch of the target repository!")
		githubactions.Errorf("Expected to find it at: %v", expectedFileUrl)

		_, _, _, fetchTargetBranchWorkflowErr := client.Repositories.GetContents(ctx, inputs.targetOwner, inputs.targetRepository, workflowFilepath, &github.RepositoryContentGetOptions{
			Ref: inputs.targetRef,
		})
		if fetchTargetBranchWorkflowErr != nil && inputs.targetRef != *targetRepository.DefaultBranch {
			// Target branch also does not include the workflow
			githubactions.Fatalf("The target workflow was also not found on the target branch: %v. Do you maybe have a typo in the filename?", inputs.targetRef)
		} else if fetchTargetBranchWorkflowErr != nil {
			// Target branch IS the default branch but the workflow is missing
			githubactions.Fatalf("Do you maybe have a typo in the filename?")
		} else if inputs.targetRef != *targetRepository.DefaultBranch {
			// Workflow is in target branch but not the default branch
			githubactions.Fatalf("Please add a dummy %v file to branch '%v' to 'register' the workflow with the GitHub API and try again!", workflowFilepath, *targetRepository.DefaultBranch)
		}

	} else {
		_, _, _, fetchTargetBranchWorkflowErr := client.Repositories.GetContents(ctx, inputs.targetOwner, inputs.targetRepository, workflowFilepath, &github.RepositoryContentGetOptions{
			Ref: inputs.targetRef,
		})
		if fetchTargetBranchWorkflowErr != nil {
			githubactions.Fatalf("The target workflow must exist at the target ref!")
		}
	}
}
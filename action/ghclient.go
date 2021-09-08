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

type GitHubClient struct {
	api                *github.Client
	apiTimeoutDuration time.Duration
	githubVars         githubVars
	inputs             inputs
}

func NewGitHubClient(githubVars githubVars, inputs inputs) *GitHubClient {
	// https://github.com/google/go-github#authentication
	// First, create an AppsTransport for initial auth
	appTransport := ghinstallation.NewAppsTransportFromPrivateKey(http.DefaultTransport, inputs.appID, inputs.privateKey)
	appTransport.BaseURL = githubVars.apiUrl

	// use appTransport to generate a client
	client := github.NewClient(&http.Client{Transport: appTransport})

	// We only need 1 installation since the app we are interested in only exists
	// in the context of its 1 installation to our org/repo
	appInstallations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{
		PerPage: 1,
	})
	if err != nil {
		githubactions.Fatalf("Error constructing new Github Client: %v", err.Error())
	}

	// Take the first (only) installation we fetched
	installationTransport := ghinstallation.NewFromAppsTransport(appTransport, *appInstallations[0].ID)

	return &GitHubClient{
		api:                github.NewClient(&http.Client{Transport: installationTransport}),
		apiTimeoutDuration: time.Second * 10,
		githubVars:         githubVars,
		inputs:             inputs,
	}
}

func (client *GitHubClient) ValidateTargetWorkflowExists(ctx context.Context) {
	workflowFilepath := fmt.Sprintf(".github/workflows/%v.yml", client.inputs.workflowFilename)
	defaultBranch := *client.GetTargetRepositoryDefaultBranch(ctx)

	workflowExistsOnDefaultBranch := client.CheckIfFileExistsAtRef(ctx, client.inputs.targetOwner, client.inputs.targetRepository, workflowFilepath, defaultBranch)
	workflowExistsOnTargetBranch := client.CheckIfFileExistsAtRef(ctx, client.inputs.targetOwner, client.inputs.targetRepository, workflowFilepath, client.inputs.targetRef)

	if !workflowExistsOnDefaultBranch || !workflowExistsOnTargetBranch {
		githubactions.Errorf("The target workflow must exist on both the default branch (%v) and target ref (%v) of the target repository!", defaultBranch, client.inputs.targetRef)
	}

	if !workflowExistsOnDefaultBranch && !workflowExistsOnTargetBranch {
		githubactions.Fatalf("No %v file was found at either ref. Perhaps you have a typo in the workflow filename?", workflowFilepath)
	} else if !workflowExistsOnDefaultBranch && workflowExistsOnTargetBranch {
		githubactions.Fatalf("Please add a dummy %v file to branch '%v' to 'register' the workflow with the GitHub API and try again!", workflowFilepath, defaultBranch)
	} else if workflowExistsOnDefaultBranch && !workflowExistsOnTargetBranch {
		githubactions.Fatalf("The workflow was found on %v but not %v!", defaultBranch, client.inputs.targetRef)
	}
}

func (client *GitHubClient) GetTargetRepositoryDefaultBranch(ctx context.Context) *string {
	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	targetRepo, _, err := client.api.Repositories.Get(apiTimeoutCtx, client.inputs.targetOwner, client.inputs.targetRepository)
	if err != nil {
		githubactions.Fatalf("Failed to fetch target repository information: %v", err.Error())
	}

	return targetRepo.DefaultBranch
}

func (client *GitHubClient) CheckIfFileExistsAtRef(ctx context.Context, owner, repository, filepath, ref string) bool {
	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	_, _, _, err := client.api.Repositories.GetContents(apiTimeoutCtx, owner, repository, filepath, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	return err == nil
}

func (client *GitHubClient) CreateCheck(ctx context.Context) *github.CheckRun {
	detailsUrl := fmt.Sprintf("%s/%s/%s/actions", client.githubVars.serverUrl, client.inputs.targetOwner, client.inputs.targetRepository)

	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	checkRun, _, err := client.api.Checks.CreateCheckRun(apiTimeoutCtx, client.githubVars.repositoryOwner, client.githubVars.repositoryName, github.CreateCheckRunOptions{
		Name:       client.inputs.workflowFilename,
		HeadSHA:    client.githubVars.sha,
		DetailsURL: &detailsUrl,
		Status:     github.String("queued"),
		StartedAt: &github.Timestamp{
			Time: time.Now(),
		},
		Output: &github.CheckRunOutput{
			Title:   github.String(client.inputs.workflowFilename),
			Summary: github.String("This report will be populated by the triggered workflow"),
		},
	})

	if err != nil {
		githubactions.Fatalf("Error creating check: %v", err.Error())
	}

	if checkRun.ID == nil {
		githubactions.Fatalf("CreateCheckRun did not return a check ID. Exiting.")
	}

	githubactions.Infof("Created new check here: %s\n", *checkRun.HTMLURL)

	return checkRun
}

func (client *GitHubClient) DispatchWorkflow(ctx context.Context, checkRun *github.CheckRun) {
	addDefaultWorkflowInputs(&client.inputs, client.githubVars, checkRun)

	fullWorkflowFilename := fmt.Sprintf("%s.yml2", client.inputs.workflowFilename)
	githubactions.Infof("Dispatching to %v workflow in %v/%v@%v\n", fullWorkflowFilename, client.inputs.targetOwner, client.inputs.targetRepository, client.inputs.targetRef)

	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	_, err := client.api.Actions.CreateWorkflowDispatchEventByFileName(apiTimeoutCtx, client.inputs.targetOwner, client.inputs.targetRepository, fullWorkflowFilename, github.CreateWorkflowDispatchEventRequest{
		Ref:    client.inputs.targetRef,
		Inputs: client.inputs.workflowInputs,
	})

	if err != nil {
		msg := fmt.Sprintf("Error disptaching event: %v", err.Error())
		client.CompleteCheckAsFailure(context.Background(), checkRun, msg)
		githubactions.Fatalf(msg)
	}
}

func (client *GitHubClient) CompleteCheckAsFailure(ctx context.Context, checkRun *github.CheckRun, reason string) {
	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	_, _, err := client.api.Checks.UpdateCheckRun(apiTimeoutCtx, client.githubVars.repositoryOwner, client.githubVars.repositoryName, *checkRun.ID, github.UpdateCheckRunOptions{
		Status: github.String("completed"),
		Conclusion: github.String("failure"),
		CompletedAt: &github.Timestamp{
			Time: time.Now(),
		},
		Output: &github.CheckRunOutput{
			Title: github.String(checkRun.Output.GetTitle()),
			Summary: github.String(reason),
		},
	})
	if err != nil {
		githubactions.Errorf(reason)
		githubactions.Fatalf("Error marking check failed after dispatch error: %v", err.Error())
	}
}

func (client *GitHubClient) FetchCheckWithRetries(ctx context.Context, checkId int64) (*github.CheckRun, error) {
	secondsBetweenAttempts := 1
	maxAttempts := 4
	attempts := 0

	for {
		check, err := client.FetchCheck(ctx, client.githubVars, checkId)
		attempts += 1

		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			githubactions.Warningf("Error fetching check %v (attempt %v of %v): %v", checkId, attempts, maxAttempts, err.Error())
			if attempts == maxAttempts {
				return nil, fmt.Errorf("Exceeded max attempts fetching check %v!", checkId)
			}

			time.Sleep(time.Second * time.Duration(secondsBetweenAttempts))
		} else {
			return check, nil
		}
	}

}

func (client *GitHubClient) FetchCheck(ctx context.Context, githubVars githubVars, checkId int64) (*github.CheckRun, error) {

	apiTimeoutCtx, cancel := context.WithTimeout(ctx, client.apiTimeoutDuration)
	defer cancel()

	check, _, err := client.api.Checks.GetCheckRun(apiTimeoutCtx, githubVars.repositoryOwner, githubVars.repositoryName, checkId)
	return check, err
}

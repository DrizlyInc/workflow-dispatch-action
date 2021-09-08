package main

import (
	"context"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	client := initializeGithubClient()

	client.ValidateTargetWorkflowExists(context.Background())

	checkRun := client.CreateCheck(context.Background())

	client.DispatchWorkflow(context.Background(), checkRun)

	waitForCheckCompletion(client, checkRun)
}

func initializeGithubClient() *GitHubClient {
	githubVars, err := parseGithubVars()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	inputs, err := parseInputs()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	client := NewGitHubClient(githubVars, inputs)

	return client
}

func waitForCheckCompletion(client *GitHubClient, checkRun *github.CheckRun) {
	if !client.inputs.waitForCheck {
		githubactions.Infof("wait_for_check was false, proceeding\n")
		return
	}

	checkTimeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(client.inputs.waitTimeoutSeconds))
	defer cancel()

	checkSucceeded, err := pollForCheckCompletion(checkTimeoutCtx, client, *checkRun.ID)
	if err != nil {
		githubactions.Fatalf("Error waiting for check to finish: %v", err.Error())
	}

	if !checkSucceeded {
		githubactions.Fatalf("Check failed!\n")
	}

	githubactions.Infof("Check completed successfully!\n")
	scrapeOutputs(client, *checkRun.ID)
}

func scrapeOutputs(client *GitHubClient, checkId int64) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	check, err := client.FetchCheckWithRetries(ctx, checkId)
	if err != nil {
		githubactions.Fatalf("Error fetching check for output scraping: %v", err.Error())
	}

	checkReportText := check.GetOutput().Text
	parsedOutputs := parseOutputsFromText(checkReportText)

	githubactions.SetOutput("output", parsedOutputs)
}

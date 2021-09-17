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

// initializeGithubClient parses environment variables (user inputs and
// standard GitHub variables) then uses those to construct and return
// a GitHub api client
func initializeGithubClient() *GitHubClient {
	githubVars, err := parseGithubVars()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	inputs, err := parseInputs()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	return NewGitHubClient(githubVars, inputs)
}

// waitForCheckCompletion waits for the given checkRun to update to a status
// of "completed" with a timeout specified as input by the user
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

// scrapeOutputs fetches the check from the repository and reads the report
// to get any outputs written as json to the end of the report. It then sets
// that json content as an output named "output" for this action
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

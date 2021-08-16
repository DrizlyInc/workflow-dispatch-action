package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	githubVars, inputs, client := initialize()
	checkRun := createCheck(client, githubVars, inputs)
	dispatchWorkflow(client, githubVars, inputs, checkRun)
	waitForCheck(client, githubVars, inputs, checkRun)
	scrapeOutputs(client, githubVars, int64(*checkRun.ID))
}

func initialize() (githubVars, inputs, *github.Client) {
	githubVars, err := parseGithubVars()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	inputs, err := parseInputs()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	client, err := newGithubClient(http.DefaultTransport, githubVars, inputs)
	if err != nil {
		githubactions.Fatalf("Error constructing new Github Client: %v", err)
	}

	return githubVars, inputs, client
}

func createCheck(client *github.Client, githubVars githubVars, inputs inputs) *github.CheckRun {
	detailsUrl := fmt.Sprintf("%s/%s/%s/actions", githubVars.serverUrl, inputs.targetOwner, inputs.targetRepository)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	checkRun, _, err := client.Checks.CreateCheckRun(ctx, githubVars.repositoryOwner, githubVars.repositoryName, github.CreateCheckRunOptions{
		Name:       inputs.workflowFilename,
		HeadSHA:    githubVars.sha,
		DetailsURL: &detailsUrl,
		Status:     github.String("queued"),
		StartedAt: &github.Timestamp{
			Time: time.Now(),
		},
		Output: &github.CheckRunOutput{
			Title:   github.String(inputs.workflowFilename),
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

func dispatchWorkflow(client *github.Client, githubVars githubVars, inputs inputs, checkRun *github.CheckRun) {
	// Add default inputs to those provided by the user
	inputs.workflowInputs["github_repository"] = githubVars.repository
	inputs.workflowInputs["github_sha"] = githubVars.sha
	inputs.workflowInputs["check_id"] = fmt.Sprint(*checkRun.ID)

	rawInputs, err := json.Marshal(inputs.workflowInputs)
	if err != nil {
		githubactions.Fatalf("Error unmarshaling workflow_inputs: %v", err.Error())
	}
	githubactions.Infof("Complete workflow inputs: %v\n", string(rawInputs))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	validateTargetWorkflowExistsOnDefaultBranch(ctx, client, githubVars, inputs)

	fullWorkflowFilename := fmt.Sprintf("%s.yml", inputs.workflowFilename)
	githubactions.Infof("Dispatching to %v workflow in %v/%v@%v\n", fullWorkflowFilename, inputs.targetOwner, inputs.targetRepository, inputs.targetRef)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

	_, err = client.Actions.CreateWorkflowDispatchEventByFileName(ctx, inputs.targetOwner, inputs.targetRepository, fullWorkflowFilename, github.CreateWorkflowDispatchEventRequest{
		Ref:    inputs.targetRef,
		Inputs: inputs.workflowInputs,
	})

	if err != nil {
		githubactions.Fatalf("Error disptaching event: %v", err.Error())
	}
}

func waitForCheck(client *github.Client, githubVars githubVars, inputs inputs, checkRun *github.CheckRun) {
	if !inputs.waitForCheck {
		githubactions.Infof("wait_for_check was false, proceeding\n")
		return
	}

	checkTimeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(inputs.waitTimeoutSeconds))
	defer cancel()

	checkSucceeded, err := pollForCheckCompletion(checkTimeoutCtx, client, githubVars, inputs, int(*checkRun.ID))
	if err != nil {
		githubactions.Fatalf("Error waiting for check to finish: %v", err.Error())
	}

	if checkSucceeded {
		githubactions.Infof("Check completed successfully!\n")
	} else {
		githubactions.Infof("Check failed!\n")
	}

}

func scrapeOutputs(client *github.Client, githubVars githubVars, checkId int64) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	check, err := fetchCheckWithRetries(ctx, client, githubVars, int(checkId))
	if err != nil {
		githubactions.Fatalf("Error fetching check for output scraping: %v", err.Error())
	}

	checkReportText := check.GetOutput().Text
	parsedOutputs := parseOutputsFromText(checkReportText)

	githubactions.SetOutput("output", parsedOutputs)
}

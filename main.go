package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	inputs, err := getInputs()
	if err != nil {
		githubactions.Fatalf("%v", err.Error())
	}

	ghrepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	owner := ghrepo[0]
	repo := ghrepo[1]

	client, err := newGithubClient(http.DefaultTransport, inputs.appID, inputs.privateKey)
	if err != nil {
		githubactions.Fatalf("Error constructing new Github Client: %v", err)
	}

	summary := fmt.Sprintf("It's running at [%s](https://github.com/%s/%s/actions)", inputs.targetRepository, inputs.targetOwner, inputs.targetRepository)

	// Create Check for the run we will trigger
	checkRun, _, err := client.Checks.CreateCheckRun(context.Background(), owner, repo, github.CreateCheckRunOptions{
		Name:    inputs.eventType,
		HeadSHA: os.Getenv("GITHUB_SHA"),
		Status:  github.String("queued"),
		StartedAt: &github.Timestamp{
			Time: time.Now(),
		},
		Output: &github.CheckRunOutput{
			Title:   github.String("Please Hold"),
			Summary: github.String(summary),
		},
	})
	if err != nil {
		githubactions.Fatalf("Error creating check: %v", err.Error())
	}
	if checkRun.ID == nil {
		githubactions.Fatalf("CreateCheckRun did not actually create a check. Exiting.")
	}
	defer githubactions.Infof("View created check here: %s", *checkRun.HTMLURL)

	githubactions.Infof("Created Check %v\n", *checkRun.ID)

	// Sets amount of time that we should wait for the GitHub API call to complete
	apiTimeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Fill out additional fields of the payload
	inputs.clientPayload["github_repository"] = fmt.Sprintf("%v/%v", owner, repo)
	inputs.clientPayload["github_sha"] = os.Getenv("GITHUB_SHA")
	inputs.clientPayload["check_id"] = fmt.Sprint(*checkRun.ID)
	rawPayload := json.RawMessage{}
	rawPayload, err = json.Marshal(inputs.clientPayload)
	if err != nil {
		githubactions.Fatalf("Error unmarshaling client payload: %v", err.Error())
	}
	githubactions.Infof("full client payload: %v\n", string(rawPayload))

	_, err = client.Actions.CreateWorkflowDispatchEventByFileName(apiTimeoutCtx, inputs.targetOwner, inputs.targetRepository, inputs.eventType, github.CreateWorkflowDispatchEventRequest{
		Ref: "br/workflow_dispatch_test",
		Inputs: inputs.clientPayload,
	})

	// _, _, err = client.Repositories.Dispatch(apiTimeoutCtx, inputs.targetOwner, inputs.targetRepository, github.DispatchRequestOptions{
	// 	EventType:     inputs.eventType,
	// 	ClientPayload: &rawPayload,
	// })
	if err != nil {
		githubactions.Fatalf("Error disptaching event: %v", err.Error())
	}

	if inputs.waitForCheck {
		// Wait for the check to be completed after all actions have passed
		checkTimeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(inputs.waitTimeoutSeconds))
		// cancel context just in case it sticks around
		defer cancel()

		// wait for our own check to finish
		err = checkWaiter(checkTimeoutCtx, client, owner, repo, int(*checkRun.ID))
		if err != nil {
			githubactions.Fatalf("Error waiting for created check to finish: %v", err.Error())
		}

		githubactions.Infof("Check passed. Moving on.\n")
	} else {
		githubactions.Infof("waitForCheck is false. Moving on.\n")
	}
}

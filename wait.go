package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

const secondsBetweenChecks = 5

func pollForCheckCompletion(ctx context.Context, client *github.Client, githubVars githubVars, inputs inputs, checkId int) (bool, error) {
	githubactions.Infof("Waiting for check %v to complete (%vs timeout) ...\n", checkId, inputs.waitTimeoutSeconds)

	// loop forever (we handle breaking out later)
	iterations := 0
	for {

		secondsRemainingUntilTimeout := inputs.waitTimeoutSeconds - int64(secondsBetweenChecks*iterations)
		githubactions.Infof("    Fetching check status (%vs remaining)... ", secondsRemainingUntilTimeout)

		apiTimeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		check, _, err := client.Checks.GetCheckRun(apiTimeoutCtx, githubVars.repositoryOwner, githubVars.repositoryName, int64(checkId))
		if err != nil {
			githubactions.Infof("FAILED\n")
			return false, fmt.Errorf("Error fetching check %v: %w", checkId, err)
		}
		githubactions.Infof("%v\n", *check.Status)

		if *check.Status == "completed" {
			return *check.Conclusion == "success", nil
		}

		// check if context has been closed(either by timeout or another error in err group).
		// If so, exit with error.
		// If not, sleep for a bit and loop again
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("Timeout: %w", ctx.Err())
		default:
			iterations += 1
			time.Sleep(time.Second * time.Duration(secondsBetweenChecks))
		}
	}
}

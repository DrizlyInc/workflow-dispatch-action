package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-githubactions"
)

const secondsBetweenChecks = 5

func pollForCheckCompletion(ctx context.Context, client *GitHubClient, checkId int64) (bool, error) {
	githubactions.Infof("Waiting for check %v to complete (%vs timeout) ...\n", checkId, client.inputs.waitTimeoutSeconds)

	// loop forever (we handle breaking out later)
	for {

		check, err := client.FetchCheckWithRetries(ctx, checkId)
		if err != nil {
			return false, fmt.Errorf("Error fetching check %v: %w", checkId, err)
		}

		secondsRemainingUntilTimeout := getSecondsRemaining(ctx)
		githubactions.Infof("    Check status (%.1fs remaining) ... %v\n", secondsRemainingUntilTimeout, *check.Status)

		if *check.Status == "completed" {
			return *check.Conclusion == "success", nil
		}

		// check if context has been closed(either by timeout or another error in err group).
		// If so, exit with error.
		// If not, sleep for a bit and loop again
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("Abandoning check waiting: %w", ctx.Err())
		default:
			time.Sleep(time.Second * time.Duration(secondsBetweenChecks))
		}
	}
}

func getSecondsRemaining(ctx context.Context) float64 {
	deadline, _ := ctx.Deadline()
	timeRemaining := deadline.Sub(time.Now())
	return timeRemaining.Seconds()
}

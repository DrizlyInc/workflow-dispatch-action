package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

func checkWaiter(ctx context.Context, client *github.Client, owner, repo string, id int) error {
	githubactions.Infof("Waiting for check %v\n", id)
	// loop forever  (we handle breaking out later)
	for {
		githubactions.Infof("Getting status of checkID %v\n", id)

		// get status of check
		apiTimeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		check, _, err := client.Checks.GetCheckRun(apiTimeoutCtx, owner, repo, int64(id))
		if err != nil {
			return fmt.Errorf("unable to get status of check %v: %w", id, err)
		}

		switch {
		case *check.Status != "completed":
			githubactions.Infof("Check %v not yet finished\n", id)
		case *check.Status == "completed" && *check.Conclusion != "success":
			return fmt.Errorf("check %v finished with unsuccessful conclusion: %v", id, *check.Conclusion)
		case *check.Status == "completed" && *check.Conclusion == "success":
			// success
			return nil
		}

		// check if context has been closed(either by timeout or another error in err group).
		// If so, exit with error.
		// If not, sleep for a bit and loop again
		select {
		case <-ctx.Done():
			return fmt.Errorf("stopping check for id %v: %w", id, ctx.Err())
		default:
			time.Sleep(time.Second * 5)
		}
	}
}

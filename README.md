# Repository Dispatch Action

This action does three things:
1. Creates a `queued` GitHub check on the repository invoking this action
2. Triggers a [`repository_dispatch` event](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#repository_dispatch) in another (target) repository
3. Optionally waits for the check to updated to a completed status by a workflow in the target repository

# Usage

```yaml
- uses: DrizlyInc/repository-dispatch-action@main
  with:

    # App ID for a GitHub app with write permissions to the dispatching repository
    # and target repository (for triggering workflows and writing creating checks)
    app_id: ${{ secrets.MY_APP_ID }}

    # Private key for the GitHub app id provided
    private_key: ${{ secrets.MY_APP_PRIVATE_KEY }}

    # Name of the repository to target with the dispatch
    target_repository: REPO_CHANGE_ME

    # Owner of the target repository (user or organization)
    target_owner: OWNER_CHANGE_ME

    # Name of the repository dispatch event type
    # this is defined in the workflow file
    event_type: EVENT_TYPE

    # If true, this action will wait until the check it creates is updated
    # to a completed status before exiting
    wait_for_check: true

    # Number of seconds to wait for the check before timing out (ignored if wait_for_check is false).
    # Inlcudes setup time to pull actions, etc
    wait_timeout_seconds: 60

    # JSON string to provide as client payload on the repository dispatch event.
    # Three additional fields are automatically added to the client_payload json object
    # prior to dispatching:
    #    check_id: The ID of the queued GitHub check created by this action
    #    GITHUB_REPOSITORY: The repository invoking this action, formatted as "<owner>/<repository-name>"
    #    GITHUB_SHA: The GITHUB_SHA in the workflow invoking this action
    client_payload: |
      {
        "variable": "foo_bar",
        "my_cool_num": 2
      }

```

# License

The contents of this repository are released under the [GNU General Public License v3.0](LICENSE)
# Workflow Dispatch Action

This action does three things:
1. Creates a `queued` GitHub check on the repository invoking this action
2. Triggers a [`workflow_dispatch` event](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_dispatch) in another (target) repository
3. Optionally waits for the check to updated to a completed status by a workflow in the target repository

# Usage

```yaml
- uses: DrizlyInc/workflow-dispatch-action@v0.1.0
  with:

    # App ID for a GitHub app with write permissions to the dispatching repository
    # and target repository (for triggering workflows and writing creating checks)
    app_id: ${{ secrets.MY_APP_ID }}

    # Private key for the GitHub app id provided
    private_key: ${{ secrets.MY_APP_PRIVATE_KEY }}

    #  Name and owner of the repository to target with the dispatch (owner/repo-name)
    target_repository: example-username/example-repository

    # Ref which should be triggered on the target repository
    target_ref: main

    # The basename (no .yml extension) of the file in .github/workflows/ of
    # the target_repository responding to the workflow_dispatch event
    workflow_filename: my-workflow

    # If true, this action will wait until the check it creates is updated
    # to a completed status before exiting
    wait_for_check: true

    # Number of seconds to wait for the check before timing out (ignored if wait_for_check is false).
    # Inlcudes setup time to pull actions, etc
    wait_timeout_seconds: 60

    # Inputs to pass to the workflow, must be a JSON encoded string ex. '{ "myinput":"myvalue" }'
    # Three additional fields are automatically added to the inputs prior to dispatching:
    #    check_id: The ID of the queued GitHub check created by this action
    #    github_repository: The repository invoking this action, formatted as "<owner>/<repository-name>"
    #    github_sha: The GITHUB_SHA in the workflow invoking this action
    workflow_inputs: |
      {
        "variable": "foo_bar",
        "my_cool_num": "2"
      }

```

# License

The contents of this repository are released under the [GNU General Public License v3.0](LICENSE)

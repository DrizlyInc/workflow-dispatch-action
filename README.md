# Workflow Dispatch Action

This action does three things:
1. Creates a `queued` GitHub check on the repository invoking this action
2. Triggers a [`workflow_dispatch` event](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_dispatch) in another (target) repository
3. Optionally waits for the check to updated to a completed status by a workflow in the target repository

# Gotchas

* `target_ref` allows you to specify which version of the workflow to trigger in the target repository, but that workflow MUST exist on the default branch in order for the GitHub API to recognize it as valid [[reference](https://docs.github.com/en/actions/managing-workflow-runs/manually-running-a-workflow#configuring-a-workflow-to-run-manually)].

# Usage

## From the Sending Workflow

### Permissions

The workflow must have access to a GitHub app with `{ contents: write, checks: write }` permissions on the source and destination repositories.


### Configuration

This example calls a remote workflow (`.github/workflows/my-workflow.yml` in the `main` branch of `example-username/example-repository`) with parameters `variable` and `my_cool_num`.  It then waits up to `60` seconds for the remote workflow to complete before continuing.

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

    # If false, this action will not wait until the check it creates is updated
    # to a completed status before exiting
    wait_for_check: false

    # Number of seconds to wait for the check before timing out (ignored if wait_for_check is false).
    # Inlcudes setup time to pull actions, etc
    wait_timeout_seconds: 120

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
      
  env:
    # Optional, can be used to inform the action which installation of the given app_id and private_key
    # to use. If not provided, the action will assume only one installation exists and use the first one
    # it finds (this should be the case in most circumstances).
    APP_INSTALLATION_ID: 18419284

```

## From the Receiving Workflow

### Permissions

The receiving workflow _must_ have access to a GitHub app with `{ contents: write, checks: write }` permissions on the source and destination repositories.


### Configuration 

Each workflow triggered by `workflow-dispatch-action` _must_ specify [`workflow_dispatch`](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_dispatch) as a triggering event.  It _must_ also declare the following inputs, in addition to any that are specific to the workflow itself.

```yaml
on:
  workflow_dispatch:
    inputs:
      check_id:
        description: The id of the check which this workflow should update
        required: true
      github_repository:
        description: The name of the repository which dispatched this workflow
        required: true
      github_sha:
        description: The sha on the repository which dispatched this workflow where the check is
        required: true     
```

In the case of the sending workflow example above, the additional inputs might look like this: 

```yaml
      variable:
        description: A random string
        required: true
      my_cool_num:
        description: an integer, represented as a string
```

### Operation

Upon execution, the workflow _should_ update the provided check (via its `check_id`) with a status of `in-progress` to indicate status back to the original repository.  Upon completion, the workflow _must_ update the provided check (via its `check_id`) with a status of `completed` (which is performed implicitly if a `conclusion` is given, which represents the successs/failure of the receiving workflow).

The receiving workflow _may_ create its own checks, recorded against the `github_repository` and `github_sha` provided as inputs to the workflow.


# Releasing

To generate a new release of this action, simply update the version tag on the image designation at the end of the [action metadata file](./action.yml). The github workflow will automatically publish a new image and create a release upon merging to main.


# License

The contents of this repository are released under the [GNU General Public License v3.0](LICENSE)

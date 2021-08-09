# repository-dispatch

This action does three things:
* Creates a Github check in the repository invoking this action
* Triggers a `repository_dispatch` [event](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#repository_dispatch) in another (target) repository
* Optionally waits for the check to be completed by the target repository

# Usage

Example Job syntax

```yaml
      - uses: DrizlyInc/repository-dispatch-action@main
        with:
          # App ID for a GitHub app with write permissions to the dispatching repository and target repository (for triggering workflows and writing creating checks)
          app_id: ${{ secrets.MY_APP_ID }}

          # Private key for the GitHub app id provided
          private_key: ${{ secrets.MY_APP_PRIVATE_KEY }}

          # Name of the repository to target with the dispatch
          target_repository: REPO_CHANGE_ME

          # Owner of the target repository (user or organization)
          target_owner: OWNER_CHANGE_ME

          # Name of the repository dispatch event type
          # this is defined in the workflow file
          event_type: PREDEFINED_EVENT_TYPE

          # "Should the action wait for its check to finish? true | false"
          wait_for_check: true

          # Number of seconds to wait for the check before timing out (ignored if wait_for_check is false).
          # Inlcudes setup time to pull actions, etc
          wait_timeout_seconds: 60

          # JSON string to provide as client payload on the repository dispatch event
          client_payload: |
            {
              "variable": "foo_bar",
              "my_cool_num": 2
            }

```
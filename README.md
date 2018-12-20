## `start-pt-story`

This command will start a Pivotal Tracker story for you and create a new
branch in whatever local repo you are currently in. That branch will have the
PT story's id added at the end of the branch name. If you've set up the
PT/GitHub integration then PT will automatically associate the branch with the
story. Later, when you make a PR for that branch, you can see the PR status in
the story.

See https://www.pivotaltracker.com/integrations/GitHub for more information
about this integration.

This command expects a config file at `$HOME/.pivotaltrackerrc` to exist. That
file must look like this:

    token = <your PT token here>
    user_id = <your PT user id here>

You can create a token in your Pivotal Tracker user profile.

You can get your `user_id` by going to
[https://www.pivotaltracker.com/services/v5/me](https://www.pivotaltracker.com/services/v5/me). It's
the `id` field in the JSON response.

The tool will look stories matching the ID you provide in all the projects to
which you belong.

This tool accepts the following options:

* -base - The base branch from which to create the new branch. (default "master")
* -branch - The new branch name (without the story ID).
* -id - The ID of the story to start.

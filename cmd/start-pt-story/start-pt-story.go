package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"code.gitea.io/git"
	"github.com/Unknwon/goconfig"
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

// Make a config file at $HOME/.pivotaltrackerrc like this:
//
// token = ...
// project_id = ...
// user_id = ...
//
// You can get your user id by going to https://www.pivotaltracker.com/services/v5/me

type config struct {
	token     string
	projectID int
	userID    int
}

func main() {
	var id string
	flag.StringVar(&id, "id", "", "The ID of the story to start")
	var base string
	flag.StringVar(&base, "base", "master", "The base branch from which to create the new branch")
	var branch string
	flag.StringVar(&branch, "branch", "", "The new branch name (without the story ID)")
	flag.Parse()

	if branch == "" {
		io.WriteString(os.Stderr, "You must provide a branch name in -branch\n")
		os.Exit(1)
	}

	config := readConfig()
	client := pivotal.NewClient(config.token)

	storyID := storyID(id)
	story, _, err := client.Stories.Get(config.projectID, storyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get story: %s\n", err)
		os.Exit(1)
	}

	checkStoryState(story, config.userID)

	if regexp.MustCompile(fmt.Sprintf("%d", storyID)).MatchString(branch) {
		fmt.Printf("You cannot put the story ID (%d) in the branch name (%s)\n", storyID, branch)
		os.Exit(1)
	}

	repo, err := git.OpenRepository(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open a git repository in the current directory: %s\n", err)
		os.Exit(1)
	}

	_, _, err = client.Stories.Update(
		config.projectID,
		storyID,
		&pivotal.StoryRequest{OwnerIds: &([]int{config.userID}), State: "started"},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not update story: %s\n", err)
		os.Exit(1)
	}

	fullBranch := fmt.Sprintf("%s-%d", branch, storyID)
	err = repo.CreateBranch(fullBranch, base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create a new branch, %s, from %s: %s\n", branch, base, err)
		os.Exit(1)
	}
	err = git.Checkout(repo.Path, git.CheckoutOptions{Branch: fullBranch})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not check out new branch, %s: %s\n", branch, err)
		os.Exit(1)
	}
}

func readConfig() config {
	p := filepath.Join(os.Getenv("HOME"), ".pivotaltrackerrc")
	c, err := goconfig.LoadConfigFile(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read INI config file at %s: %s\n", p, err)
	}

	token, err := c.GetValue("DEFAULT", "token")
	if err != nil {
		io.WriteString(os.Stderr, "Could not get token value from config data\n")
		os.Exit(1)
	}
	projectID, err := c.Int64("DEFAULT", "project_id")
	if err != nil {
		io.WriteString(os.Stderr, "Could not get project_id value from config data\n")
		os.Exit(1)
	}
	userID, err := c.Int64("DEFAULT", "user_id")
	if err != nil {
		io.WriteString(os.Stderr, "Could not get user_id value from config data\n")
		os.Exit(1)
	}

	return config{token: token, userID: int(userID), projectID: int(projectID)}
}

func storyID(id string) int {
	if id == "" {
		io.WriteString(os.Stderr, "You must provide a story id in -id\n")
		os.Exit(1)
	}

	i, err := strconv.ParseInt(regexp.MustCompile(`^#`).ReplaceAllLiteralString(id, ""), 10, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not turn %s into an integer: %s\n", id, err)
		os.Exit(1)
	}

	return int(i)
}

func checkStoryState(story *pivotal.Story, userID int) {
	if story.State == "started" {
		for _, id := range story.OwnerIds {
			if id == userID {
				return
			}
		}
	}

	if story.State != "unstarted" {
		fmt.Fprintf(os.Stderr, "The story is in an unexpected state: %s\n", story.State)
		os.Exit(1)
	}
	if len(story.OwnerIds) > 0 {
		io.WriteString(os.Stderr, "The story already has an owner\n")
		os.Exit(1)
	}
}

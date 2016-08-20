package main

import (
	"fmt"
	"strings"
)

var force bool
var defaults bool

func cmdInit(args []string) {
	var masterBranch string

	flags.BoolVar(&force, "f", false,
		"force setting of git-drip branches, even if already configured")
	flags.BoolVar(&defaults, "d", false, "use default branch naming conventions")
	flags.Parse(args)

	err := runErrNoOutput("git", "rev-parse", "--git-dir")
	if err != nil {
		run("git", "init")
	} else if !gitRepoHeadless() {
		requireCleanTree()
	}

	if initializedForGitFlow() {
		dief("Already initialized for gitflow")
	}

	if gitDripInitialized() && !force {
		dief("Already initialized for gitdrip.\nTo force reinitialization, " +
			"use: git drip init -f")
	}

	if defaults {
		fmt.Fprintln(stdout(), "Using default branch names.")
	}

	if isDripMasterConfigured() && !force {
		masterBranch = config()["gitdrip.branch.master"]
	} else {
		// Two cases are distinguished:
		// 1. A fresh git repo (without any branches)
		//    We will configure repo for user
		// 2. Some branches do already exist
		//    We will disallow creation of new branches and
		//    rather allow to use existing branches for git-drip.
		var suggestion, answer string
		var shouldCheck, ok bool
		var branches = LocalBranches()

		if len(branches) == 0 {
			fmt.Fprintln(stdout(), "No branches exist yet. Base branches must be created now.")
			if suggestion, ok = config()["gitdrip.branch.master"]; !ok {
				suggestion = "master"
			}
		} else {
			fmt.Fprintln(stdout(), "\nWhich branch should be used for development?")
			for _, b := range branches {
				fmt.Fprintf(stdout(), "   - %s\n", b.Name)
			}
			shouldCheck = true
			if suggestion, ok = config()["gitdrip.branch.master"]; !ok || !BranchesContains(branches, suggestion) {
				suggestion = "master"
			}
		}

		fmt.Fprintf(stdout(), "Branch name for development: [%s] ", suggestion)
		if defaults {
			fmt.Fprintln(stdout(), "")
			masterBranch = suggestion
		} else {
			fmt.Scanln(&answer)
			if strings.TrimSpace(answer) == "" {
				masterBranch = suggestion
			} else {
				masterBranch = answer
			}
		}

		// check existence in case of an already existing repo
		if shouldCheck {
			if !BranchesContains(branches, masterBranch) {
				dief("Local branch '%s' does not exist.", masterBranch)
			}
		}

		setConfiguration("gitdrip.branch.master", masterBranch)
	}

	// Creation of HEAD
	// ----------------
	// We create a HEAD now, if it does not exist yet (in a fresh repo). We need
	// it to be able to create new branches.
	if gitRepoHeadless() {
		run("git", "symbolic-ref", "HEAD", fmt.Sprintf("refs/heads/%s", masterBranch))
		run("git", "commit", "--allow-empty", "--quiet", "-m", "Initial commit")
		run("git", "checkout", "-q", masterBranch)
	}

	// finally, ask the user for naming conventions (branch and tag prefixes)
	config := config()
	if force || !hasKey(config, "gitdrip.prefix.feature") ||
		!hasKey(config, "gitdrip.prefix.release") ||
		!hasKey(config, "gitdrip.prefix.hotfix") ||
		!hasKey(config, "gitdrip.prefix.versiontag") {
		var suggestion, prefix, answer string
		var ok bool

		fmt.Fprintln(stdout(), "\nHow to name your supporting branch prefixes?")

		for _, ps := range []struct {
			key        string
			suggestion string
			question   string
		}{
			{
				"gitdrip.prefix.feature",
				"feature/",
				"Feature branches? [%s] ",
			},
			{
				"gitdrip.prefix.release",
				"release/",
				"Release branches? [%s] ",
			},
			{
				"gitdrip.prefix.hotfix",
				"hotfix/",
				"Hotfix branches? [%s] ",
			},
			{
				"gitdrip.prefix.versiontag",
				"",
				"Version tag prefix? [%s] ",
			},
		} {

			if suggestion, ok = config[ps.key]; !ok {
				suggestion = ps.suggestion
			}
			fmt.Fprintf(stdout(), ps.question, suggestion)
			if defaults {
				fmt.Fprintln(stdout(), "")
				prefix = suggestion
			} else {
				fmt.Scanln(&answer)
				if strings.TrimSpace(answer) == "" {
					prefix = suggestion
				} else {
					prefix = answer
				}
			}

			run("git", "config", ps.key, prefix)
		}
	}

	fmt.Fprintln(stdout(), "\ngit drip has been initialized")
}

func hasKey(c map[string]string, key string) bool {
	_, ok := c[key]
	return ok
}

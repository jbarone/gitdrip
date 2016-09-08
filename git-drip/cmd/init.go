// Copyright © 2016 Joshua Barone
//
// This file is part of git-drip.
//
// git-drip is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// git-drip is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with git-drip. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [-fd]",
	Short: "Initialize a new git repo with support for the workflow",
	Long: `Initialize a new git repo with support for the workflow.

Initialization will prompt you with questions and options to help setup
your environment to work with the git-drip workflow.`,
	Run: InitRepo,
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolP("force", "f", false,
		"Force setting of git-drip branches, even if already configured")
	initCmd.Flags().BoolP("defaults", "d", false,
		"Use default branch naming conventions")
}

// InitRepo initializes the git repo for git-drip workflow
func InitRepo(cmd *cobra.Command, args []string) {

	if len(args) > 0 {
		_ = cmd.Help() // #nosec error is ignored since exiting
		os.Exit(1)
	}

	force, _ := cmd.Flags().GetBool("force")       // #nosec
	defaults, _ := cmd.Flags().GetBool("defaults") // #nosec

	checkCleanGitRepo()

	if isFlowInitialized() {
		dief("Already initialized for gitflow")
	}

	if isDripInitialized() && !force {
		dief("Already initialized for git-drip.\nTo force reinitialization, " +
			"use: git drip init -f")
	}

	if defaults {
		fmt.Fprintln(stdout(), "Using default branch names.")
	}

	var masterBranch = configureDevelop(force, defaults)
	enforceHead(masterBranch)

	configurePrefixes(force, defaults)

	fmt.Fprintln(stdout(), "\ngit drip has been initialized")
}

func checkCleanGitRepo() {

	_, err := cmdOutputErr("git", "rev-parse", "--git-dir")
	if err != nil {
		run("git", "init")
	}

	if !RepoHeadless() {
		RequireCleanTree()
	}
}

func getDevelopSuggestion(force bool) (string, bool) {
	var suggestion string
	var shouldCheck, ok bool
	var branches = LocalBranches()

	if len(branches) == 0 { // fresh repo
		fmt.Fprintln(stdout(),
			"No branches exist yet. Base branches must be created now.")
		if suggestion, ok = GitConfig()["gitdrip.branch.master"]; !ok {
			suggestion = "master"
		}
	} else {
		fmt.Fprintln(stdout(),
			"\nWhich branch should be used for development?")
		for _, b := range branches {
			fmt.Fprintf(stdout(), "   - %s\n", b.Name)
		}
		shouldCheck = true
		suggestion, ok = GitConfig()["gitdrip.branch.master"]
		if !ok || !BranchesContains(branches, suggestion) {
			suggestion = "master"
		}
	}
	return suggestion, shouldCheck
}

func configureDevelop(force, defaults bool) string {
	if isDripMasterConfigured() && !force {
		return GitConfig()["gitdrip.branch.master"]
	}

	var answer, masterBranch string

	suggestion, shouldCheck := getDevelopSuggestion(force)

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
	if shouldCheck && !BranchesContains(LocalBranches(), masterBranch) {
		dief("Local branch '%s' does not exist.", masterBranch)
	}

	setConfiguration("gitdrip.branch.master", masterBranch)

	return masterBranch
}

func enforceHead(masterBranch string) {
	if RepoHeadless() {
		run("git", "symbolic-ref", "HEAD",
			fmt.Sprintf("refs/heads/%s", masterBranch))
		run("git", "commit", "--allow-empty", "--quiet",
			"-m", "Initial commit")
		run("git", "checkout", "-q", masterBranch)
	}
}

func configurePrefixes(force, defaults bool) {
	config := GitConfig()
	if force || !hasKey(config, "gitdrip.prefix.feature") ||
		!hasKey(config, "gitdrip.prefix.release") ||
		!hasKey(config, "gitdrip.prefix.hotfix") ||
		!hasKey(config, "gitdrip.prefix.versiontag") {
		var suggestion, prefix, answer string
		var ok bool

		fmt.Fprintln(stdout(), "\nHow to name supporting branch prefixes?")

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
}

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

package gitdrip

import (
	"fmt"
	"strings"

	"github.com/renstrom/dedent"
)

func getBranchNameWidth(branches []*Branch) (width int) {
	for _, b := range branches {
		if len(b.Name) > width {
			width = len(b.Name)
		}
	}
	return
}

func getFeatureBranch(name string) *Branch {
	branch := &Branch{
		Name:   name,
		Prefix: Config().Get(dripFeature),
	}
	branches := PrefixedBranches(Config().Get(dripFeature))
	if contains(branches, branch.PrefixedName()) {
		return branch
	}

	var matches []*Branch
	for _, b := range branches {
		if strings.HasPrefix(b.PrefixedName(), branch.PrefixedName()) {
			matches = append(matches, b)
		}
	}

	switch len(matches) {
	case 0:
		dief("No branch matches prefix " + branch.Name)
	case 1:
		return matches[0]
	default:
		fmt.Fprintf(stderr(), "Multiple branches match prefix '%s':",
			branch.Name)
		for _, m := range matches {
			fmt.Fprintln(stderr(), "-", m.FullName())
		}
		die()
	}
	return nil
}

func getFeatureBranchOrCurrent(arg string) *Branch {
	if arg != "" {
		return getFeatureBranch(arg)
	}

	branch := CurrentBranch()
	if branch == nil || branch.Prefix != Config().Get(dripFeature) {
		dief("The current HEAD is not a feature branch.\n" +
			"Please specify a <name> argument")
	}

	return branch
}

func finishFeatureCleanup(branch *Branch, origin, master string,
	remote, keep bool) {
	requireBranch(branch)
	requireCleanTree()

	if remote {
		run("git", "push", origin, ":"+branch.FullName())
	}

	if !keep {
		run("git", "branch", "-d", branch.PrefixedName())
	}

	// print summary
	fmt.Fprintln(stdout(), "\nSummary of actions:")
	fmt.Fprintf(stdout(), "- The feature branch '%s' was merged into '%s'\n",
		branch.PrefixedName(), master)
	switch keep {
	case true:
		fmt.Fprintf(stdout(),
			"- Feature branch '%s' is still available\n", branch.PrefixedName())
	case false:
		fmt.Fprintf(stdout(),
			"- Feature branch '%s' has been removed\n", branch.PrefixedName())
	}
	fmt.Fprintf(stdout(),
		"- You are now on branch '%s'\n\n", master)
}

// ListFeatures displays the feature branches for the repo
func ListFeatures(descriptions bool) {
	var prefix = Config().Get(dripFeature)
	var masterBranch = Config().Get(dripMaster)
	var featureBranches = PrefixedBranches(prefix)
	var currentBranch = CurrentBranch()

	if len(featureBranches) == 0 {
		fmt.Fprintln(stderr(), dedent.Dedent(`No feature branches exists.

		You can start a new feature branch:

		git drip feature start <name> [<base]
		`))
		return
	}

	width := getBranchNameWidth(featureBranches) + 3

	for _, b := range featureBranches {
		if b.FullName() == currentBranch.Name {
			fmt.Fprintf(stdout(), "* ")
		} else {
			fmt.Fprintf(stdout(), "  ")
		}

		var description, extra string
		if descriptions || verbose > 0 {
			description, _ = trimErr(cmdOutputErr("git", "config",
				"branch."+b.PrefixedName()+".description"))
			description += " "
		}

		if verbose > 0 {
			base := cmdOutput("git", "merge-base", b.FullName(), masterBranch)
			developSha := cmdOutput("git", "rev-parse", masterBranch)
			branchSha := cmdOutput("git", "rev-parse", b.FullName())
			extra = "(may be rebased)"
			switch {
			case branchSha == developSha:
				extra = "(no commits yet)"
			case base == branchSha:
				extra = "(is behind develop, may ff)"
			case base == developSha:
				extra = "(based on latest develop)"
			}
		}
		fmt.Fprintf(stdout(), fmt.Sprintf("%%-%ds%%s%%s\n", width),
			b.Name, description, extra)
	}
}

// DescribeFeature displays the feature branches for the repo
func DescribeFeature(brancharg, description string) {
	branch := getFeatureBranchOrCurrent(brancharg)

	if description != "" {
		Config().Set(
			fmt.Sprintf("branch.%s.description", branch.PrefixedName()),
			description)
	} else {
		run("git", "branch", "--edit-description", branch.PrefixedName())
	}

	// print summary
	fmt.Fprintf(stdout(), dedent.Dedent(`
	Summary of actions:

	- The local branch '%s' had description edited

	`), branch.PrefixedName())
}

// StartFeatures creates a new feature branch
func StartFeatures(branchname, basearg, message string, fetch, describe bool) {
	branch := &Branch{
		Name:   branchname,
		Prefix: Config().Get(dripFeature),
	}

	requireBranchAbsent(branch)

	master := Config().Get(dripMaster)
	origin := "origin"
	if Config().Has(dripOrigin) {
		origin = Config().Get(dripOrigin)
	}

	base := master
	if basearg != "" {
		base = basearg
	}

	if fetch {
		run("git", "fetch", "-q", origin, master)
	}

	if remoteContains(origin + "/" + master) {
		requireEqual(master, origin+"/"+master)
	}

	err := runErr("git", "checkout", "-b", branch.FullName(), base)
	if err != nil {
		dief("Could not create feature branch '%s'", branch.FullName())
	}

	if message != "" {
		Config().Set(
			fmt.Sprintf("branch.%s.description", branch.FullName()), message)
	}

	if describe {
		run("git", "branch", "--edit-description", branch.FullName())
	}

	// print summary
	fmt.Fprintf(stdout(), dedent.Dedent(`
		Summary of actions:
		- A new branch '%s' was created, based on '%s'
		- You are now on branch '%s'

		Now, start committing on your feature. When done, use:
			 git drip feature finish %s

		`),
		branch.PrefixedName(),
		base,
		branch.PrefixedName(),
		branch.Name)
}

// DeleteFeature deletes a given feature branch
func DeleteFeature(brancharg string, remote bool) {
	branch := getFeatureBranch(brancharg)
	master := Config().Get(dripMaster)
	requireBranch(branch)
	requireCleanTree()

	run("git", "checkout", master)

	if remote {
		origin := "origin"
		if Config().Has(dripOrigin) {
			origin = Config().Get(dripOrigin)
		}
		run("git", "push", origin, ":"+branch.PrefixedName())
	}

	run("git", "branch", "-d", branch.PrefixedName())

	fmt.Fprintf(stdout(), dedent.Dedent(`
	Summary of actions:
	- Feature branch '%s' has been removed
	- You are now on branch '%s'

	`), branch.PrefixedName(), master)
}

// CheckoutFeature checks out the specified feature branch
func CheckoutFeature(brancharg string) {
	run("git", "checkout", getFeatureBranch(brancharg).PrefixedName())
}

// DiffFeature displays the diff date of the feature branch
func DiffFeature(brancharg string) {
	if brancharg != "" {
		branch := getFeatureBranch(brancharg)
		base := trim(cmdOutput("git", "merge-base",
			Config().Get(dripMaster), branch.PrefixedName()))
		run("git", "diff", fmt.Sprintf("%s..%s", base, branch.PrefixedName()))
		return
	}

	branch := CurrentBranch()
	if branch == nil || branch.Prefix != Config().Get(dripFeature) {
		dief("Not on a feature branch. Name one explicitly.")
	}
	base := trim(cmdOutput("git", "merge-base",
		Config().Get(dripMaster), "HEAD"))
	run("git", "diff", base)
}

// RebaseFeature rebases the feature branch on master
func RebaseFeature(brancharg string, interactive bool) {
	branch := getFeatureBranchOrCurrent(brancharg)
	var opts string
	if interactive {
		opts = "-i"
	}

	printf("Will try to rebase '%s'", branch.Name)

	requireCleanTree()
	requireBranch(branch)

	run("git", "checkout", "-q", branch.PrefixedName())
	run("git", "rebase", opts, Config().Get(dripMaster))
}

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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// featureCmd represents the feature command
var featureCmd = &cobra.Command{
	Use:   "feature",
	Short: "Manage your feature branches",
	Long:  `Manage your feature branches`,
	Run:   ListFeatures,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		requireDripInit()
	},
}

// featureListCmd represents the feature list command
var featureListCmd = &cobra.Command{
	Use:   "list [-d]",
	Short: "List feature branches",
	Long:  "List feature branches",
	Run:   ListFeatures,
}

// featureStartCmd represents the feature list command
var featureStartCmd = &cobra.Command{
	Use:   `start [-Fd] [-m "<description>"] <name> [<base>]`,
	Short: "Start feature branches",
	Long:  "Start feature branches",
	Run:   StartFeatures,
}

// featureDescribeCmd represents the feature describe command
var featureDescribeCmd = &cobra.Command{
	Use:   `describe [-m "<descirption>"] [<name|nameprefix>]`,
	Short: "Add description to feature branch",
	Long:  "Add description to feature branch",
	Run:   DescribeFeature,
}

// featureFinishCmd represents the feature finish command
var featureFinishCmd = &cobra.Command{
	Use:   "finish [-rFks] [<name|nameprefix>]",
	Short: "Finish feature branch",
	Long:  "Finish feature branch",
	Run:   FinishFeature,
}

// featureDeleteCmd represents the feature delete command
var featureDeleteCmd = &cobra.Command{
	Use:   "delete [-F] <name|nameprefix>",
	Short: "Delete feature branch",
	Long:  "Delete feature branch",
	Run:   DeleteFeature,
}

// featureCheckoutCmd represents the feature checkout|co commands
var featureCheckoutCmd = &cobra.Command{
	Use:     "checkout <name|nameprefix>",
	Aliases: []string{"co"},
	Short:   "Checkout feature branch",
	Long:    "Checkout feature branch",
	Run:     CheckoutFeature,
}

// featureDiffCmd represents the feature diff command
var featureDiffCmd = &cobra.Command{
	Use:   "diff [<name|nameprefix>]",
	Short: "Show diff for feature branch",
	Long:  "Show diff for feature branch",
	Run:   DiffFeature,
}

// featureRebaseCmd represents the feature rebase command
var featureRebaseCmd = &cobra.Command{
	Use:   "rebase [-i] [<name|nameprefix>]",
	Short: "Rebase feature branch",
	Long:  "REbase feature branch",
	Run:   RebaseFeature,
}

// featurePublishCmd represents the feature publish command
var featurePublishCmd = &cobra.Command{
	Use:   "publish [<name|nameprefix>]",
	Short: "Publish feature branch",
	Long:  "Publish feature branch",
	Run:   PublishFeature,
}

// featureTrackCmd represents the feature track command
var featureTrackCmd = &cobra.Command{
	Use:   "track [<name|nameprefix>]",
	Short: "Track feature branch",
	Long:  "Track feature branch",
	Run:   TrackFeature,
}

// featurePullCmd represents the feature pull command
var featurePullCmd = &cobra.Command{
	Use:   "pull <remote> [<name|nameprefix>]",
	Short: "Pull feature branch",
	Long:  "Pull feature branch",
	Run:   PullFeature,
}

func init() {
	// feature root
	RootCmd.AddCommand(featureCmd)
	featureCmd.Flags().BoolP("descriptions", "d", false,
		"Include branch descriptions")
	_ = viper.BindPFlag("descriptions",
		featureCmd.Flags().Lookup("descriptions"))

	// list
	featureCmd.AddCommand(featureListCmd)
	featureListCmd.Flags().BoolP("descriptions", "d", false,
		"Include branch descriptions")
	_ = viper.BindPFlag("descriptions",
		featureListCmd.Flags().Lookup("descriptions"))
	viper.SetDefault("descriptions", false)

	// start
	featureCmd.AddCommand(featureStartCmd)
	featureStartCmd.Flags().BoolP("fetch", "F", false,
		"Fetch from remote origin before performing local operations")
	featureStartCmd.Flags().BoolP("describe", "d", false,
		"Open editor to add description for feature branch")
	featureStartCmd.Flags().StringP("message", "m", "",
		"Add a given description for feature branch")

	// describe
	featureCmd.AddCommand(featureDescribeCmd)
	featureDescribeCmd.Flags().StringP("message", "m", "",
		"Add a given description for feature branch")

	// delete
	featureCmd.AddCommand(featureDeleteCmd)
	featureDeleteCmd.Flags().BoolP("remote", "R", false,
		"Delete remote branch as well")

	// finish
	featureCmd.AddCommand(featureFinishCmd)
	featureFinishCmd.Flags().BoolP("fetch", "F", false,
		"Fetch from remote origin before performing finish")
	featureFinishCmd.Flags().BoolP("rebase", "r", false,
		"Rebase branch instead of merge")
	featureFinishCmd.Flags().BoolP("keep", "k", false,
		"Keep branch after performing finish")
	featureFinishCmd.Flags().BoolP("squash", "S", false,
		"Squash feature while performing merge")

	// checkout
	featureCmd.AddCommand(featureCheckoutCmd)

	// diff
	featureCmd.AddCommand(featureDiffCmd)

	// rebase
	featureCmd.AddCommand(featureRebaseCmd)
	featureRebaseCmd.Flags().BoolP("interactive", "i", false,
		"Do an interactive rebase")

	// publish
	featureCmd.AddCommand(featurePublishCmd)

	// track
	featureCmd.AddCommand(featureTrackCmd)

	// pull
	featureCmd.AddCommand(featurePullCmd)
}

func getBranchNameWidth(branches []*Branch) (width int) {
	for _, b := range branches {
		if len(b.Name) > width {
			width = len(b.Name)
		}
	}
	return
}

func featurePrefix() string {
	return GitConfig()["gitdrip.prefix.feature"]
}

func getFeatureBranch(prefix, name string) *Branch {
	branch := &Branch{
		Name:   name,
		Prefix: prefix,
	}
	branches := LocalBranches()
	if BranchesContains(branches, branch.FullName()) {
		return branch
	}

	var matches []*Branch
	for _, b := range branches {
		if strings.HasPrefix(b.FullName(), branch.FullName()) {
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

func getCurrentBranch(prefix string) *Branch {
	branch := CurrentBranch()

	if strings.HasPrefix(branch.Name, prefix) {
		branch.Prefix = prefix
		branch.Name = strings.TrimPrefix(branch.Name, prefix)
		return branch
	}

	return nil
}

func getFeatureBranchOrCurrent(args []string) *Branch {
	prefix := featurePrefix()
	if len(args) > 0 {
		return getFeatureBranch(prefix, args[0])
	}

	branch := getCurrentBranch(prefix)
	if branch == nil {
		dief("The current HEAD is not a feature branch.\n" +
			"Please specify a <name> argument")
	}

	return branch
}

func finishFeatureCleanup(branch *Branch, origin, master string,
	remote, keep bool) {
	RequireBranch(branch)
	RequireCleanTree()

	if remote {
		run("git", "push", origin, ":refs/heads/"+branch.FullName())
	}

	if !keep {
		run("git", "branch", "-d", branch.FullName())
	}

	// print summary
	fmt.Fprintln(stdout(), "\nSummary of actions:")
	fmt.Fprintf(stdout(), "- The feature branch '%s' was merged into '%s'\n",
		branch.FullName(), master)
	switch keep {
	case true:
		fmt.Fprintf(stdout(),
			"- Feature branch '%s' is still available\n", branch.FullName())
	case false:
		fmt.Fprintf(stdout(),
			"- Feature branch '%s' has been removed\n", branch.FullName())
	}
	fmt.Fprintf(stdout(),
		"- You are now on branch '%s'\n\n", master)
}

// ListFeatures displays the feature branches for the repo
func ListFeatures(cmd *cobra.Command, args []string) {
	descriptions := viper.GetBool("descriptions")
	// descriptions, _ := cmd.Flags().GetBool("descriptions") // #nosec

	var config = GitConfig()
	var prefix = config["gitdrip.prefix.feature"]
	var masterBranch = config["gitdrip.branch.master"]
	var featureBranches = PrefixedBranches(prefix)
	var currentBranch = CurrentBranch()

	if len(featureBranches) == 0 {
		fmt.Fprintln(stderr(), "No feature branches exists.")
		fmt.Fprintln(stderr(), "")
		fmt.Fprintln(stderr(), "You can start a new feature branch:")
		fmt.Fprintln(stderr(), "")
		fmt.Fprintln(stderr(), "    git drip feature start <name> [<base>]")
		fmt.Fprintln(stderr(), "")
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
		if descriptions || *verbose > 0 {
			description = trim(cmdOutput("git", "config",
				"branch."+b.FullName()+".description")) + " "
		}

		if *verbose > 0 {
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

// StartFeatures displays the feature branches for the repo
func StartFeatures(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Missing argument <name>")
	}

	branch := &Branch{
		Name:   args[0],
		Prefix: featurePrefix(),
	}

	fetch, _ := cmd.Flags().GetBool("fetch")          // #nosec
	describe, _ := cmd.Flags().GetBool("description") // #nosec
	message, _ := cmd.Flags().GetString("message")    // #nosec

	RequireBranchAbsent(branch)

	config := GitConfig()
	master := config["gitdrip.branch.master"]
	origin := "origin"
	if hasKey(config, "gitdrip.origin") {
		origin = config["gitdrip.origin"]
	}

	base := master
	if len(args) == 2 {
		base = args[1]
	}

	if fetch {
		run("git", "fetch", "-q", origin, master)
	}

	if BranchesContains(RemoteBranches(), origin+"/"+master) {
		RequireEqual(master, origin+"/"+master)
	}

	err := runErr("git", "checkout", "-b", branch.FullName(), base)
	if err != nil {
		dief("Could not create feature branch '%s'", branch.FullName())
	}

	if message != "" {
		setConfiguration(
			fmt.Sprintf("branch.%s.description", branch.FullName()), message)
	}

	if describe {
		run("git", "branch", "--edit-description", branch.FullName())
	}

	// print summary
	fmt.Fprintln(stdout(), "\nSummary of actions:")
	fmt.Fprintf(stdout(), "- A new branch '%s' was created, based on '%s'\n",
		branch.FullName(), base)
	fmt.Fprintf(stdout(),
		"- You are now on branch '%s'\n\n", branch.FullName())
	fmt.Fprintln(stdout(),
		"Now, start committing on your feature. When done, use:")
	fmt.Fprintf(stdout(), "     git drip feature finish %s\n\n", branch.Name)
}

// DescribeFeature displays the feature branches for the repo
func DescribeFeature(cmd *cobra.Command, args []string) {
	branch := getFeatureBranchOrCurrent(args)
	description, _ := cmd.Flags().GetString("message") // #nosec

	if description != "" {
		setConfiguration(
			fmt.Sprintf("branch.%s.description", branch.FullName()),
			description)
	} else {
		run("git", "branch", "--edit-description", branch.FullName())
	}

	// print summary
	fmt.Fprintln(stdout(), "\nSummary of actions:")
	fmt.Fprintf(stdout(),
		"\n- The local branch '%s' had description edited\n\n",
		branch.FullName())
}

// DeleteFeature ...
func DeleteFeature(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Missing argument <name|nameprefix>")
	}
	branch := getFeatureBranch(featurePrefix(), args[0])
	config := GitConfig()
	master := config["gitdrip.branch.master"]
	RequireBranch(branch)
	RequireCleanTree()

	run("git", "checkout", master)

	remote, _ := cmd.Flags().GetBool("remote") // #nosec
	if remote {
		run("git", "push", origin(config), ":refs/heads/"+branch.FullName())
	}

	run("git", "branch", "-d", branch.FullName())

	fmt.Fprintln(stdout(), "\nSummary of actions:")
	fmt.Fprintf(stdout(), "- Feature branch '%s' has been removed\n",
		branch.FullName())
	fmt.Fprintf(stdout(), "- You are now on branch '%s'\n\n", master)
}

func featureResolveMerge(branch *Branch, path, master string,
	remote, keep bool) {
	if gitWorkingTreeStatus() != clean {
		fmt.Fprintln(stdout(), "\nMerge conflicts not resolved yet, use:")
		fmt.Fprintln(stdout(), "    git mergetool")
		fmt.Fprintln(stdout(), "    git commit")
		fmt.Fprintln(stdout(),
			"\nYou can then complete the finish by running it again:")
		fmt.Fprintf(stdout(),
			"    git drip feature finish %s\n\n", branch.Name)
		die()
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		dief(err.Error())
	}

	_ = os.Remove(path) // #nosec
	finishBase := trim(string(content))
	if branch.isMergedInto(finishBase) {
		finishFeatureCleanup(branch, master, origin(GitConfig()), remote, keep)
		os.Exit(0)
	}
}

// FinishFeature ...
func FinishFeature(cmd *cobra.Command, args []string) {
	branch := getFeatureBranchOrCurrent(args)
	RequireBranch(branch)
	remote, _ := cmd.Flags().GetBool("remote")
	keep, _ := cmd.Flags().GetBool("keep")
	squash, _ := cmd.Flags().GetBool("squash")
	rebase, _ := cmd.Flags().GetBool("rebase")
	config := GitConfig()
	master := config["gitdrip.branch.master"]

	path := filepath.Join(gitDir(), ".gitdrip", "MERGE_BASE")
	if ok, _ := exists(path); ok {
		// restoring from merge conflict
		featureResolveMerge(branch, path, master, remote, keep)
	}

	RequireCleanTree()
	remoteBranch := origin(config) + "/" + branch.FullName()
	if BranchesContains(RemoteBranches(), remoteBranch) {
		if remote {
			run("git", "fetch", "-q", origin(config), branch.FullName())
		}
		RequireEqual(master, remoteBranch)
	}
	remoteMaster := origin(config) + "/" + master
	if BranchesContains(RemoteBranches(), remoteMaster) {
		RequireEqual(master, remoteMaster)
	}

	if rebase {
		err := runErr("git", "drip", "feature", "rebase",
			branch.Name, remoteMaster)
		if err != nil {
			fmt.Fprintln(stderr(),
				"Finish was aborted due to conflicts during rebase.")
			fmt.Fprintln(stderr(), "Please finish the rebase manually now.")
			fmt.Fprintln(stderr(), "When finished, re-run")
			fmt.Fprintf(stderr(), "    git drip feature finish '%s' '%s'\n",
				branch.FullName(), master)
		}
	}

	run("git", "checkout", master)
	var opts string
	if squash {
		opts = "--squash"
	}
	err := runErr("git", "merge", opts, branch.FullName())
	if err != nil {
		_ = os.MkdirAll(filepath.Dir(path), 0644)        // #nosec
		_ = ioutil.WriteFile(path, []byte(master), 0644) // #nosec
		fmt.Fprintln(stdout(),
			"\nThere were merge conflicts. To resolve the merge conflict "+
				"manually, use:")
		fmt.Fprintln(stdout(), "    git mergetool")
		fmt.Fprintln(stdout(), "    git commit")
		fmt.Fprintln(stdout(),
			"\nYou can then complete the finish by running it again:")
		fmt.Fprintf(stdout(),
			"    git drip feature finish %s\n\n", branch.Name)
		die()
	}

	finishFeatureCleanup(branch, master, origin(config), remote, keep)
}

// CheckoutFeature checks out the specified feature branch
func CheckoutFeature(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Name a feature branch explicitly.")
	}
	branch := getFeatureBranch(featurePrefix(), args[0])
	run("git", "checkout", branch.FullName())
}

// DiffFeature displays the diff date of the feature branch
func DiffFeature(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		branch := getFeatureBranch(featurePrefix(), args[0])
		base := trim(cmdOutput("git", "merge-base",
			GitConfig()["gitdrip.branch.master"], branch.FullName()))
		run("git", "diff", fmt.Sprintf("%s..%s", base, branch.FullName()))
		return
	}

	branch := getCurrentBranch(featurePrefix())
	if branch == nil {
		dief("Not on a feature branch. Name one explicitly.")
	}
	base := trim(cmdOutput("git", "merge-base",
		GitConfig()["gitdrip.branch.master"], "HEAD"))
	run("git", "diff", base)
}

// RebaseFeature rebases the feature branch on master
func RebaseFeature(cmd *cobra.Command, args []string) {
	branch := getFeatureBranchOrCurrent(args)

	interactive, _ := cmd.Flags().GetBool("interactive") // #nosec
	var opts string
	if interactive {
		opts = "-i"
	}

	printf("Will try to rebase '%s'", branch.Name)

	RequireCleanTree()
	RequireBranch(branch)

	run("git", "checkout", "-q", branch.FullName())
	run("git", "rebase", opts, GitConfig()["gitdrip.branch.master"])
}

// PublishFeature ...
func PublishFeature(cmd *cobra.Command, args []string) {
}

// TrackFeature ...
func TrackFeature(cmd *cobra.Command, args []string) {
}

// PullFeature ...
func PullFeature(cmd *cobra.Command, args []string) {
}

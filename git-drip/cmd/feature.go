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
	"github.com/jbarone/gitdrip"
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
		cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		gitdrip.RequireDripInitialized()
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

// ListFeatures displays the feature branches for the repo
func ListFeatures(cmd *cobra.Command, args []string) {
	descriptions := viper.GetBool("descriptions")
	gitdrip.ListFeatures(descriptions)
}

// StartFeatures displays the feature branches for the repo
func StartFeatures(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Missing argument <name>")
	}

	basearg := ""
	if len(args) > 1 {
		basearg = args[1]
	}

	fetch, _ := cmd.Flags().GetBool("fetch")          // #nosec
	describe, _ := cmd.Flags().GetBool("description") // #nosec
	message, _ := cmd.Flags().GetString("message")    // #nosec

	gitdrip.StartFeatures(args[0], basearg, message, fetch, describe)
}

// DescribeFeature displays the feature branches for the repo
func DescribeFeature(cmd *cobra.Command, args []string) {
	description, _ := cmd.Flags().GetString("message") // #nosec

	brancharg := ""
	if len(args) > 0 {
		brancharg = args[0]
	}

	gitdrip.DescribeFeature(brancharg, description)
}

// DeleteFeature ...
func DeleteFeature(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Missing argument <name|nameprefix>")
	}
	remote, _ := cmd.Flags().GetBool("remote") // #nosec

	gitdrip.DeleteFeature(args[0], remote)
}

// FinishFeature ...
func FinishFeature(cmd *cobra.Command, args []string) {
	brancharg := ""
	if len(args) > 0 {
		brancharg = args[0]
	}
	remote, _ := cmd.Flags().GetBool("remote")
	keep, _ := cmd.Flags().GetBool("keep")
	squash, _ := cmd.Flags().GetBool("squash")
	rebase, _ := cmd.Flags().GetBool("rebase")

	gitdrip.FinishFeature(brancharg, remote, keep, squash, rebase)
}

// CheckoutFeature checks out the specified feature branch
func CheckoutFeature(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		dief("Name a feature branch explicitly.")
	}
	gitdrip.CheckoutFeature(args[0])
}

// DiffFeature displays the diff date of the feature branch
func DiffFeature(cmd *cobra.Command, args []string) {
	brancharg := ""
	if len(args) > 0 {
		brancharg = args[0]
	}
	gitdrip.DiffFeature(brancharg)
}

// RebaseFeature rebases the feature branch on master
func RebaseFeature(cmd *cobra.Command, args []string) {
	brancharg := ""
	if len(args) > 0 {
		brancharg = args[0]
	}
	interactive, _ := cmd.Flags().GetBool("interactive") // #nosec
	gitdrip.RebaseFeature(brancharg, interactive)
}

// PublishFeature ...
func PublishFeature(cmd *cobra.Command, args []string) {
	brancharg := ""
	if len(args) > 0 {
		brancharg = args[0]
	}
	gitdrip.PublishFeature(brancharg)
}

// TrackFeature ...
func TrackFeature(cmd *cobra.Command, args []string) {
}

// PullFeature ...
func PullFeature(cmd *cobra.Command, args []string) {
}

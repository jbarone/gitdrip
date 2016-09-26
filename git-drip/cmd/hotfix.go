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

// hotfixCmd represents the hotfix command
var hotfixCmd = &cobra.Command{
	Use:   "hotfix",
	Short: "Manage your hotfix branches",
	Long:  "Manage your hotfix branches",
	Run:   ListHotfixes,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		gitdrip.RequireDripInitialized()
	},
}

// hotfixListCmd represents the hotfix list command
var hotfixListCmd = &cobra.Command{
	Use:   "list [-d]",
	Short: "List hotfix branches",
	Long:  "List hotfix branches",
	Run:   ListHotfixes,
}

func init() {
	RootCmd.AddCommand(hotfixCmd)

	hotfixCmd.Flags().BoolP("descriptions", "d", false,
		"Include branch descriptions")
	_ = viper.BindPFlag("hotfix.descriptions",
		hotfixCmd.Flags().Lookup("descriptions"))

	// list
	hotfixCmd.AddCommand(hotfixListCmd)
	hotfixListCmd.Flags().BoolP("descriptions", "d", false,
		"Include branch descriptions")
	_ = viper.BindPFlag("hotfix.descriptions",
		hotfixListCmd.Flags().Lookup("descriptions"))
	viper.SetDefault("hotfix.descriptions", false)
}

// ListHotfixes displays the feature branches for the repo
func ListHotfixes(cmd *cobra.Command, args []string) {
	descriptions := viper.GetBool("hotfix.descriptions")
	gitdrip.ListHotfixes(descriptions)
}

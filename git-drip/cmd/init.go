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
	"os"

	"github.com/jbarone/gitdrip"
	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [-fd]",
	Short: "Initialize a new git repo with support for the workflow",
	Long: dedent.Dedent(`
		Initialize a new git repo with support for the workflow.

		Initialization will prompt you with questions and options to help setup
		your environment to work with the git-drip workflow.`),
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

	gitdrip.InitDrip(force, defaults)
}

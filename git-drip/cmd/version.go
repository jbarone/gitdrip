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
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const gitDripVersion = "0.4.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(stdout(), "%s\n", gitDripVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

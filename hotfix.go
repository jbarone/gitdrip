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

	"github.com/renstrom/dedent"
)

// ListHotfixes displays the hotfix branches for the repo
func ListHotfixes(descriptions bool) {
	var prefix = Config().Get(dripHotfix)
	var masterBranch = Config().Get(dripMaster)
	var hotfixBranches = PrefixedBranches(prefix)
	var currentBranch = CurrentBranch()

	if len(hotfixBranches) == 0 {
		fmt.Fprintln(stderr(), dedent.Dedent(`
		No hotfix branches exists.

		You can start a new hotfix branch:

		git drip hotfix start <name> [<base]
		`))
		return
	}

	width := getBranchNameWidth(hotfixBranches) + 3

	for _, b := range hotfixBranches {
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
			base := trim(cmdOutput("git", "merge-base",
				b.FullName(), masterBranch))
			developSha := cmdOutput("git", "rev-parse", masterBranch)
			branchSha := cmdOutput("git", "rev-parse", b.FullName())
			tagname, err := trimErr(cmdOutputErr("git", "name-rev", "--tags",
				"--no-undefined", "--name-only", base))
			switch {
			case branchSha == developSha:
				extra = "(no commits yet)"
			case err == nil:
				extra = fmt.Sprintf("(based on %s)", tagname)
			default:
				extra = fmt.Sprintf("(based on %s)",
					trim(cmdOutput("git", "rev-parse", "--short", base)))
			}
		}
		fmt.Fprintf(stdout(), fmt.Sprintf("%%-%ds%%s%%s\n", width),
			b.Name, description, extra)
	}
}

// DescribeHotfix displays the feature branches for the repo
func DescribeHotfix(brancharg, description string) {
	describeBranch(Config().Get(dripHotfix), brancharg, description)
}

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

package gitdrip

import "testing"

func TestInitRequire(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	write(t, gt.client+"/file", "new content")
	trun(t, gt.client, "git", "add", "file")
	trun(t, gt.client, "git", "commit", "-m", "msg")
	write(t, gt.client+"/file", "more new content")

	t.Log("unstaged")
	InitDrip(false, true)
	testPrintedStderr(t, "fatal: Working tree contains unstaged changes. Aborting")

	trun(t, gt.client, "git", "add", "file")

	t.Log("uncommited")
	InitDrip(false, true)
	testPrintedStderr(t, "fatal: Working tree contains uncommited changes. Aborting")
}

func TestInitGitDripAlready(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	stderrTrap = testStderr
	trun(t, gt.client, "git", "config", "gitdrip.branch.master", "master")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.feature", "feature/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.release", "release/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.hotfix", "hotfix/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.versiontag", "")
	trun(t, gt.client, "git", "commit", "--allow-empty", "-m", "msg")

	InitDrip(false, true)
	testPrintedStderr(t, "Already initialized for git-drip.",
		"To force reinitialization, use: git drip init -f")
	stderrTrap = nil
}

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

import (
	"bytes"
	"testing"
)

func testInit(t *testing.T, canDie, force, defaults bool) {
	// reset git vars
	gitconfig = nil
	gitdir = ""
	gitroot = ""

	defer testCleanup(t, canDie)

	dieTrap = func() {
		died = true
		panic("died")
	}
	died = false
	runLogTrap = []string{} // non-nil, to trigger saving of commands
	stdoutTrap = new(bytes.Buffer)
	stderrTrap = new(bytes.Buffer)

	InitDrip(force, defaults)
}

func TestInitRequire(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	write(t, gt.client+"/file", "new content")
	trun(t, gt.client, "git", "add", "file")
	trun(t, gt.client, "git", "commit", "-m", "msg")
	write(t, gt.client+"/file", "more new content")

	t.Log("unstaged")
	testInit(t, true, false, true)
	testPrintedStderr(t,
		"fatal: Working tree contains unstaged changes. Aborting")

	trun(t, gt.client, "git", "add", "file")

	t.Log("uncommited")
	testInit(t, true, false, true)
	testPrintedStderr(t,
		"fatal: Working tree contains uncommited changes. Aborting")
}

func TestInitGitDripAlready(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	trun(t, gt.client, "git", "config", "gitdrip.branch.master", "master")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.feature", "feature/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.release", "release/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.hotfix", "hotfix/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.versiontag", "")
	trun(t, gt.client, "git", "commit", "--allow-empty", "-m", "msg")

	testInit(t, true, false, true)
	testPrintedStderr(t,
		"Already initialized for git-drip.",
		"To force reinitialization, use: git drip init -f")
}

func TestInitEmpty(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	trun(t, gt.client, "rm", "-rf", ".git")

	testInit(t, false, false, true)
	testRan(t,
		"git init",
		"git config gitdrip.branch.master master",
		"git symbolic-ref HEAD refs/heads/master",
		"git commit --allow-empty --quiet -m Initial commit",
		"git checkout -q master",
		"git config gitdrip.prefix.feature feature/",
		"git config gitdrip.prefix.release release/",
		"git config gitdrip.prefix.hotfix hotfix/",
		"git config gitdrip.prefix.versiontag",
	)
}

func TestInitDrip(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	testInit(t, false, false, true)
	testRan(t,
		"git config gitdrip.branch.master master",
		"git config gitdrip.prefix.feature feature/",
		"git config gitdrip.prefix.release release/",
		"git config gitdrip.prefix.hotfix hotfix/",
		"git config gitdrip.prefix.versiontag",
	)
}

func TestInitDripForce(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	trun(t, gt.client, "git", "config", "gitdrip.branch.master", "master")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.feature", "feature/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.release", "release/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.hotfix", "hotfix/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.versiontag", "")
	trun(t, gt.client, "git", "commit", "--allow-empty", "-m", "msg")

	testInit(t, false, true, true)
	testRan(t,
		"git config gitdrip.branch.master master",
		"git config gitdrip.prefix.feature feature/",
		"git config gitdrip.prefix.release release/",
		"git config gitdrip.prefix.hotfix hotfix/",
		"git config gitdrip.prefix.versiontag",
	)
}

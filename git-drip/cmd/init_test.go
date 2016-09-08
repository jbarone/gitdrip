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

import "testing"

func TestInitEmpty(t *testing.T) {
	gt := newGitTestFolder(t)
	defer gt.done()

	testMain(t, "init")
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

func TestInitRequire(t *testing.T) {
	gt := newGitTestInit(t)
	defer gt.done()

	write(t, gt.client+"/file", "new content")
	trun(t, gt.client, "git", "add", "file")
	trun(t, gt.client, "git", "commit", "-m", "msg")
	write(t, gt.client+"/file", "more new content")

	t.Log("unstaged")
	testMainDied(t, "init")
	testPrintedStderr(t, "fatal: Working tree contains unstaged changes. Aborting")

	trun(t, gt.client, "git", "add", "file")

	t.Log("uncommited")
	testMainDied(t, "init")
	testPrintedStderr(t, "fatal: Working tree contains uncommited changes. Aborting")
}

func TestInitGitFlow(t *testing.T) {
	gt := newGitTestInit(t)
	defer gt.done()

	trun(t, gt.client, "git", "config", "gitflow.branch.master", "master")
	trun(t, gt.client, "git", "config", "gitflow.branch.develop", "develop")
	trun(t, gt.client, "git", "config", "gitflow.prefix.feature", "feature/")
	trun(t, gt.client, "git", "config", "gitflow.prefix.release", "release/")
	trun(t, gt.client, "git", "config", "gitflow.prefix.hotfix", "hotfix/")
	trun(t, gt.client, "git", "config", "gitflow.prefix.support", "support/")
	trun(t, gt.client, "git", "config", "gitflow.prefix.versiontag", "")

	testMainDied(t, "init")
	testPrintedStderr(t, "Already initialized for gitflow\n")
}

func TestInitGitDrip(t *testing.T) {
	gt := newGitTestInit(t)
	defer gt.done()

	trun(t, gt.client, "git", "config", "gitdrip.branch.master", "master")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.feature", "feature/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.release", "release/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.hotfix", "hotfix/")
	trun(t, gt.client, "git", "config", "gitdrip.prefix.versiontag", "")
	write(t, gt.client+"/file", "new content")
	trun(t, gt.client, "git", "add", "file")
	trun(t, gt.client, "git", "commit", "-m", "msg")

	testMainDied(t, "init")
	testPrintedStderr(t, "Already initialized for git-drip.",
		"To force reinitialization, use: git drip init -f")
}

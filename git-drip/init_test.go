package main

import "testing"

func TestInitEmpty(t *testing.T) {
	gt := newGitTestFolder(t)
	defer gt.done()

	testMain(t, "init")
	testRan(t,
		"git rev-parse --git-dir",
		"git init",
		"git config gitdrip.branch.master master",
		"git rev-parse --quiet --verify HEAD",
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
	testPrintedStderr(t, "Already initialized for gitdrip.",
		"To force reinitialization, use: git drip init -f")
}

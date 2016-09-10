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

func TestGitConfig(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	runLogTrap = []string{} // non-nil, to trigger saving of commands
	c := Config()
	if c.Get("user.name") != "gopher" {
		t.Errorf("config(user.name)=%s expected gopher", c.Get("user.name"))
	}
	if c.Get("user.email") != "gopher@example.com" {
		t.Errorf("config(user.email)=%s expected gopher@example.com",
			c.Get("user.email"))
	}
	if c.Get("gitdrip.madeup") != "" {
		t.Errorf("config(gitdrip.madeup)=%s expected ''",
			c.Get("gitdrip.madeup"))
	}

	runLog = runLogTrap
	testRan(t)
}

func TestGitConfigSet(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	runLogTrap = []string{} // non-nil, to trigger saving of commands
	c := Config()
	if c.Get("gitdrip.madeup") != "" {
		t.Errorf("config(gitdrip.madeup)=%s expected ''",
			c.Get("gitdrip.madeup"))
	}
	c.Set("gitdrip.madeup", "newvalue")
	if c.Get("gitdrip.madeup") != "newvalue" {
		t.Errorf("config(gitdrip.madeup)=%s expected 'newvalue'",
			c.Get("gitdrip.madeup"))
	}

	runLog = runLogTrap
	testRan(t, "git config gitdrip.madeup newvalue")
}

func TestGitDir(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	path := GitDir()
	if path != ".git" {
		t.Fatalf("Expected: %s got: %s", ".git", path)
	}
}

func TestGitRoot(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	path := GitRoot()
	if path != "." {
		t.Fatalf("Expected: %s got: %s", ".", path)
	}
}

func TestCurrentBranch(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	t.Logf("on master")
	checkCurrentBranch(t, "master", "origin/master", false, false, "", "")

	t.Logf("on newbranch")
	trun(t, gt.client, "git", "checkout", "--no-track", "-b", "newbranch")
	checkCurrentBranch(t, "newbranch", "origin/master", true, false, "", "")

	t.Logf("making change")
	write(t, gt.client+"/file", "i made a change")
	trun(t, gt.client, "git", "commit", "-a", "-m", "My change line.\n\nChange-Id: I0123456789abcdef0123456789abcdef\n")
	checkCurrentBranch(t, "newbranch", "origin/master", true, true, "I0123456789abcdef0123456789abcdef", "My change line.")

	t.Logf("on dev.branch")
	trun(t, gt.client, "git", "checkout", "-t", "-b", "dev.branch", "origin/dev.branch")
	checkCurrentBranch(t, "dev.branch", "origin/dev.branch", false, false, "", "")

	t.Logf("on newdev")
	trun(t, gt.client, "git", "checkout", "-t", "-b", "newdev", "origin/dev.branch")
	checkCurrentBranch(t, "newdev", "origin/dev.branch", true, false, "", "")

	t.Logf("making change")
	write(t, gt.client+"/file", "i made another change")
	trun(t, gt.client, "git", "commit", "-a", "-m", "My other change line.\n\nChange-Id: I1123456789abcdef0123456789abcdef\n")
	checkCurrentBranch(t, "newdev", "origin/dev.branch", true, true, "I1123456789abcdef0123456789abcdef", "My other change line.")

	t.Logf("detached head mode")
	trun(t, gt.client, "git", "checkout", "HEAD^0")
	checkCurrentBranch(t, "HEAD", "origin/HEAD", false, false, "", "")
}

func checkCurrentBranch(t *testing.T, name, origin string, isLocal, hasPending bool, changeID, subject string) {
	b := CurrentBranch()
	if b.Name != name {
		t.Errorf("b.Name = %q, want %q", b.Name, name)
	}
	if x := b.OriginBranch(); x != origin {
		t.Errorf("b.OriginBranch() = %q, want %q", x, origin)
	}
	if x := b.IsLocalOnly(); x != isLocal {
		t.Errorf("b.IsLocalOnly() = %v, want %v", x, isLocal)
	}
	if x := b.HasPendingCommit(); x != hasPending {
		t.Errorf("b.HasPendingCommit() = %v, want %v", x, isLocal)
	}
	if work := b.Pending(); len(work) > 0 {
		c := work[0]
		if x := c.ChangeID; x != changeID {
			t.Errorf("b.Pending()[0].ChangeID = %q, want %q", x, changeID)
		}
		if x := c.Subject; x != subject {
			t.Errorf("b.Pending()[0].Subject = %q, want %q", x, subject)
		}
	}
}

func TestHasStagedChanges(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	t.Logf("clean repo")
	if HasStagedChanges() {
		t.Fatal("Has staged changes, but expected none")
	}

	t.Logf("making change")
	write(t, gt.client+"/file", "i made a change")
	if HasStagedChanges() {
		t.Fatal("Has staged changes, but expected none")
	}

	t.Logf("adding file")
	trun(t, gt.client, "git", "add", gt.client+"/file")
	if !HasStagedChanges() {
		t.Fatal("Has no staged changes, but expected one")
	}
}

func TestHasUnstagedChanges(t *testing.T) {
	gt := NewGitTest(t)
	defer gt.Done()

	if HasUnstagedChanges() {
		t.Fatal("Has unstaged changes, but expected none")
	}

	t.Logf("making change")
	write(t, gt.client+"/file", "i made a change")
	if !HasUnstagedChanges() {
		t.Fatal("Has no unstaged changes, but expected one")
	}

	t.Logf("adding file")
	trun(t, gt.client, "git", "add", gt.client+"/file")
	if HasUnstagedChanges() {
		t.Fatal("Has unstaged changes, but expected none")
	}
}

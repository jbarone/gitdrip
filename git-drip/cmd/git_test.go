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
	"reflect"
	"testing"
)

func TestLocalBranches(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	t.Logf("on master")
	checkLocalBranches(t, "master")

	t.Logf("on dev branch")
	trun(t, gt.client, "git", "checkout", "-b", "newbranch")
	checkLocalBranches(t, "master", "newbranch")

	t.Logf("detached head mode")
	trun(t, gt.client, "git", "checkout", "HEAD^0")
	checkLocalBranches(t, "HEAD", "master", "newbranch")
}

func checkLocalBranches(t *testing.T, want ...string) {
	var names []string
	branches := LocalBranches()
	for _, b := range branches {
		names = append(names, b.Name)
	}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("LocalBranches() = %v, want %v", names, want)
	}
}

func TestBranchesContains(t *testing.T) {
	branches := []*Branch{
		&Branch{Name: "test"},
	}

	t.Logf("test")
	if !BranchesContains(branches, "test") {
		t.Error("BranchesContains() = false, want true")
	}

	t.Logf("master")
	if BranchesContains(branches, "master") {
		t.Error("BranchesContains() = true, want false")
	}
}

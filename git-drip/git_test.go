package main

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

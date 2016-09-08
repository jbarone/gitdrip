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

import "strings"

// Branch describes a Git branch.
type Branch struct {
	Name   string
	Prefix string
}

// CurrentBranch returns the current branch.
func CurrentBranch() *Branch {
	out, err := cmdOutputErr("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil
	}
	name := strings.TrimPrefix(trim(out), "heads/")
	return &Branch{Name: name}
}

// DetachedHead reports whether branch b corresponds to a detached HEAD
// (does not have a real branch name).
func (b *Branch) DetachedHead() bool {
	return b.Name == "HEAD"
}

// FullName returns the full name of branch
func (b *Branch) FullName() string {
	return b.Prefix + b.Name
}

func (b *Branch) isMergedInto(base string) bool {
	for _, s := range nonBlankLines(
		cmdOutput("git", "branch", "--no-color", "--contains", b.FullName())) {
		s = trim(s)
		if strings.HasPrefix(s, "* ") {
			s = strings.TrimPrefix(s, "* ")
		}
		if s == base {
			return true
		}
	}
	return false
}

// RemoteBranches returns a list of all known remote branches.
func RemoteBranches() []*Branch {
	var branches []*Branch
	for _, s := range nonBlankLines(
		cmdOutput("git", "branch", "-r", "-q", "--no-color")) {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "* ") {
			s = strings.TrimPrefix(s, "* ")
		}
		branches = append(branches, &Branch{Name: s})
	}
	return branches
}

// LocalBranches returns a list of all known local branches.
// If the current directory is in detached HEAD mode, one returned
// branch will have Name == "HEAD" and DetachedHead() == true.
func LocalBranches() []*Branch {
	var branches []*Branch
	current := CurrentBranch()
	for _, s := range nonBlankLines(
		cmdOutput("git", "branch", "-q", "--no-color")) {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "* ") {
			// * marks current branch in output.
			// Normally the current branch has a name like any other,
			// but in detached HEAD mode the branch listing shows
			// a localized (translated) textual description instead of
			// a branch name. Avoid language-specific differences
			// by using CurrentBranch().Name for the current branch.
			// It detects detached HEAD mode in a more portable way.
			// (git rev-parse --abbrev-ref HEAD returns 'HEAD').
			if current != nil {
				s = current.Name
			} else {
				s = strings.TrimPrefix(s, "* ")
			}
		}
		branches = append(branches, &Branch{Name: s})
	}
	return branches
}

// PrefixedBranches returns a list of all known local branches with specified
// prefix.
func PrefixedBranches(prefix string) []*Branch {
	var branches []*Branch
	current := CurrentBranch()
	for _, s := range nonBlankLines(
		cmdOutput("git", "branch", "-q", "--no-color")) {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "* ") {
			// * marks current branch in output.
			// Normally the current branch has a name like any other,
			// but in detached HEAD mode the branch listing shows
			// a localized (translated) textual description instead of
			// a branch name. Avoid language-specific differences
			// by using CurrentBranch().Name for the current branch.
			// It detects detached HEAD mode in a more portable way.
			// (git rev-parse --abbrev-ref HEAD returns 'HEAD').
			if current != nil {
				s = current.Name
			} else {
				s = strings.TrimPrefix(s, "* ")
			}
		}

		if strings.HasPrefix(s, prefix) {
			branches = append(branches, &Branch{
				Name:   strings.TrimPrefix(s, prefix),
				Prefix: prefix,
			})
		}
	}
	return branches
}

// BranchesContains checks if the list of branches contains the one specified
func BranchesContains(branches []*Branch, name string) bool {
	for _, b := range branches {
		if b.Name == name {
			return true
		}
	}
	return false
}

type compareStatus int

const (
	equal     compareStatus = iota
	behind    compareStatus = iota
	ahead     compareStatus = iota
	needMerge compareStatus = iota
	noBase    compareStatus = iota
)

// compareBranches compares two given branches
func compareBranches(local, remote string) compareStatus {
	commit1 := trim(cmdOutput("git", "rev-parse", local))
	commit2 := trim(cmdOutput("git", "rev-parse", remote))
	if commit1 == commit2 {
		return equal
	}

	base, err := trimErr(cmdOutputErr("git", "merge-base", commit1, commit2))
	if err != nil {
		return noBase
	}

	switch {
	case commit1 == base:
		return behind
	case commit2 == base:
		return ahead
	}

	return needMerge
}

// RequireEqual dies if the given branches are not equal;
// or if local is ahead of remote a warning is displayed but execution
// continues
func RequireEqual(local, remote string) {
	stat := compareBranches(local, remote)
	if stat == equal {
		return
	}

	printf("Branches '%s' and '%s' have divereged.", local, remote)

	switch stat {
	case behind:
		dief("And branch '%s' may be fast-forwarded.", local)
	case ahead:
		printf("And local branch '%s' is ahead of '%s'.", local, remote)
	default:
		dief("Branches need merging first.")
	}
}

// RepoHeadless checks if the current git repo is headless
func RepoHeadless() bool {
	_, err := cmdOutputErr("git", "rev-parse", "--quiet", "--verify", "HEAD")
	return err != nil
}

type workingTreeStatus int

const (
	clean      workingTreeStatus = iota
	unstaged   workingTreeStatus = iota
	uncommited workingTreeStatus = iota
)

// gitWorkingTreeStatus returns the current status of the git working tree
func gitWorkingTreeStatus() workingTreeStatus {
	if _, err := cmdOutputErr("git", "diff", "--no-ext-diff",
		"--ignore-submodules", "--quiet", "--exit-code"); err != nil {
		return unstaged
	}
	if _, err := cmdOutputErr("git", "diff-index", "--cached", "--quiet",
		"--ignore-submodules", "HEAD", "--"); err != nil {
		return uncommited
	}
	return clean
}

// RequireCleanTree will die if the working tree is not in a clean state
func RequireCleanTree() {
	switch gitWorkingTreeStatus() {
	case unstaged:
		dief("fatal: Working tree contains unstaged changes. Aborting.")
	case uncommited:
		dief("fatal: Working tree contains uncommited changes. Aborting.")
	}
}

// RequireBranch dies if the requested branch doesn't exist
func RequireBranch(branch *Branch) {
	if !BranchesContains(LocalBranches(), branch.FullName()) {
		dief("Branch '%s' does not exist and is required", branch.FullName())
	}
}

// RequireBranchAbsent dies if the requested branch exists
func RequireBranchAbsent(branch *Branch) {
	if BranchesContains(LocalBranches(), branch.FullName()) {
		dief("Branch '%s' already exists. Pick another name",
			branch.FullName())
	}
}

// GitConfig returns the configuration of the git repo
func GitConfig() map[string]string {
	cfg := make(map[string]string)

	lines := nonBlankLines(cmdOutput("git", "config", "--list"))
	for _, line := range lines {
		parts := strings.Split(line, "=")
		cfg[parts[0]] = strings.Join(parts[1:], "=")
	}

	return cfg
}

func setConfiguration(key, val string) {
	run("git", "config", key, val)
}

func isFlowInitialized() bool {
	config := GitConfig()

	master, ok := config["gitflow.branch.master"]
	if !ok || master == "" {
		return false
	}
	develop, ok := config["gitflow.branch.develop"]
	if !ok || develop == "" {
		return false
	}
	if master == develop {
		return false
	}
	for _, prefix := range []string{
		"gitflow.prefix.feature",
		"gitflow.prefix.release",
		"gitflow.prefix.hotfix",
		"gitflow.prefix.support",
		"gitflow.prefix.versiontag",
	} {
		if _, ok := config[prefix]; !ok {
			return false
		}
	}

	return true
}

func isDripMasterConfigured() bool {
	config := GitConfig()

	master, ok := config["gitdrip.branch.master"]
	if !ok || master == "" || !BranchesContains(LocalBranches(), master) {
		return false
	}
	return true
}

func areDripPrefixesConfigured() bool {
	config := GitConfig()

	for _, prefix := range []string{
		"gitdrip.prefix.feature",
		"gitdrip.prefix.release",
		"gitdrip.prefix.hotfix",
		"gitdrip.prefix.versiontag",
	} {
		if _, ok := config[prefix]; !ok {
			return false
		}
	}

	return true
}

func isDripInitialized() bool {
	return isDripMasterConfigured() && areDripPrefixesConfigured()
}

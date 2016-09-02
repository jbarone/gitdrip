package main

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

// LocalBranches returns a list of all known local branches.
// If the current directory is in detached HEAD mode, one returned
// branch will have Name == "HEAD" and DetachedHead() == true.
func LocalBranches() []*Branch {
	var branches []*Branch
	current := CurrentBranch()
	for _, s := range nonBlankLines(cmdOutput("git", "branch", "-q", "--no-color")) {
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

// BranchesContains checks if the list of branches contains the one specified
func BranchesContains(branches []*Branch, name string) bool {
	for _, b := range branches {
		if b.Name == name {
			return true
		}
	}
	return false
}

func gitRepoHeadless() bool {
	err := runErrNoOutput("git", "rev-parse", "--quiet", "--verify", "HEAD")
	return err != nil
}

type workingTreeStatus int

const (
	clean      workingTreeStatus = iota
	unstaged   workingTreeStatus = iota
	uncommited workingTreeStatus = iota
)

func gitWorkingTreeStatus() workingTreeStatus {
	if err := runErrNoOutput("git", "diff", "--no-ext-diff", "--ignore-submodules", "--quiet", "--exit-code"); err != nil {
		return unstaged
	}
	if err := runErrNoOutput("git", "diff-index", "--cached", "--quiet", "--ignore-submodules", "HEAD", "--"); err != nil {
		return uncommited
	}
	return clean
}

func requireCleanTree() {
	switch gitWorkingTreeStatus() {
	case unstaged:
		dief("fatal: Working tree contains unstaged changes. Aborting.")
	case uncommited:
		dief("fatal: Working tree contains uncommited changes. Aborting.")
	}
}

func config() map[string]string {
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

func initializedForGitFlow() bool {
	config := config()

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
	config := config()

	master, ok := config["gitdrip.branch.master"]
	if !ok || master == "" || !BranchesContains(LocalBranches(), master) {
		return false
	}
	return true
}

func areDripPrefixConfigured() bool {
	config := config()

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

func gitDripInitialized() bool {
	return isDripMasterConfigured() && areDripPrefixConfigured()
}

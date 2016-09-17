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
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// HEAD is the name of the HEAD branch
const HEAD = "HEAD"

// GitConfig stores the current state of git configuration
type GitConfig struct {
	cfg map[string]string
}

var (
	gitdir    string
	gitroot   string
	gitconfig *GitConfig
)

// Config returns the current state of git configuration
func Config() *GitConfig {
	if gitconfig == nil {
		gitconfig = &GitConfig{}
		gitconfig.cfg = make(map[string]string)

		lines := nonBlankLines(cmdOutput("git", "config", "--list"))
		for _, line := range lines {
			parts := strings.Split(line, "=")
			switch len(parts) {
			case 1:
				gitconfig.cfg[parts[0]] = ""
			default:
				gitconfig.cfg[parts[0]] = strings.Join(parts[1:], "=")
			}
		}
	}
	return gitconfig
}

// Get the value for the given key
func (c *GitConfig) Get(key string) string {
	v, ok := c.cfg[key]
	if !ok {
		return ""
	}
	return v
}

// Set the given key to the given value in the git configuration
func (c *GitConfig) Set(key, value string) {
	run("git", "config", key, value)
	c.cfg[key] = value
}

// Has returns if the given key is in the configuration
func (c *GitConfig) Has(key string) bool {
	_, ok := c.cfg[key]
	return ok
}

// GitDir returns the path to the git folder for this project
func GitDir() string {
	if gitdir == "" {
		var err error
		gitdir, err = trimErr(cmdOutputErr("git", "rev-parse", "--git-dir"))
		if err != nil {
			gitdir = ""
		}
	}
	return gitdir
}

// GitRoot returns the path to the root of the git project
func GitRoot() string {
	if gitroot == "" {
		dir := GitDir()
		if dir != "" {
			gitroot, _ = filepath.Split(dir)
		}
		if gitroot == "" {
			gitroot = "."
		}
	}
	return gitroot
}

var stagedRE = regexp.MustCompile(`^[ACDMR]  `)

// HasStagedChanges reports whether the working directory contains staged changes.
func HasStagedChanges() bool {
	for _, s := range nonBlankLines(cmdOutput("git", "status", "-b", "--porcelain")) {
		if stagedRE.MatchString(s) {
			return true
		}
	}
	return false
}

var unstagedRE = regexp.MustCompile(`^.[ACDMR]`)

// HasUnstagedChanges reports whether the working directory contains unstaged changes.
func HasUnstagedChanges() bool {
	for _, s := range nonBlankLines(cmdOutput("git", "status", "-b", "--porcelain")) {
		if unstagedRE.MatchString(s) {
			return true
		}
	}
	return false
}

// Branch describes a Git branch.
type Branch struct {
	Name          string    // branch name
	Prefix        string    // git-drip prefix for branch
	loadedPending bool      // following fields are valid
	originBranch  string    // upstream origin branch
	commitsAhead  int       // number of commits ahead of origin branch
	commitsBehind int       // number of commits behind origin branch
	branchpoint   string    // latest commit hash shared with origin branch
	pending       []*Commit // pending commits, newest first (children before parents)
}

// A Commit describes a single pending commit on a Git branch.
type Commit struct {
	Hash      string // commit hash
	ShortHash string // abbreviated commit hash
	Parent    string // parent hash
	Merge     string // for merges, hash of commit being merged into Parent
	Message   string // commit message
	Subject   string // first line of commit message
	ChangeID  string // Change-Id in commit message ("" if missing)
}

// CurrentBranch returns the current branch.
func CurrentBranch() *Branch {
	name, err := trimErr(cmdOutputErr("git", "rev-parse", "--abbrev-ref", HEAD))
	if err != nil {
		return nil
	}
	name = strings.TrimPrefix(name, "heads/")
	prefix, name := branchPrefix(name)
	return &Branch{Name: name, Prefix: prefix}
}

// DetachedHead reports whether branch b corresponds to a detached HEAD
// (does not have a real branch name).
func (b *Branch) DetachedHead() bool {
	return b.Name == HEAD
}

// OriginBranch returns the name of the origin branch that branch b tracks.
// The returned name is like "origin/master" or "origin/dev.garbage" or
// "origin/release-branch.go1.4".
func (b *Branch) OriginBranch() string {
	if b.DetachedHead() {
		// Detached head mode.
		// "origin/HEAD" is clearly false, but it should be easy to find when it
		// appears in other commands. Really any caller of OriginBranch
		// should check for detached head mode.
		return "origin/HEAD"
	}

	if b.originBranch != "" {
		return b.originBranch
	}
	argv := []string{"git", "rev-parse", "--abbrev-ref", b.Name + "@{u}"}
	out, err := exec.Command(argv[0], argv[1:]...).CombinedOutput() // #nosec
	if err == nil && len(out) > 0 {
		b.originBranch = string(bytes.TrimSpace(out))
		return b.originBranch
	}

	// Have seen both "No upstream configured" and "no upstream configured".
	if strings.Contains(string(out), "upstream configured") {
		// Assume branch was created before we set upstream correctly.
		b.originBranch = "origin/master"
		return b.originBranch
	}
	fmt.Fprintf(stderr(), "%v\n%s\n", commandString(argv[0], argv[1:]), out)
	dief("%v", err)
	panic("not reached")
}

// FullName of branch
func (b *Branch) FullName() string {
	if b.Name != HEAD {
		return "refs/heads/" + b.PrefixedName()
	}
	return b.Name
}

// PrefixedName returns the name of the branch with the prefix
func (b *Branch) PrefixedName() string {
	if b.Name == HEAD || b.Prefix == "" {
		return b.Name
	}
	return b.Prefix + b.Name
}

// IsLocalOnly reports whether b is a local work branch (only local, not known to remote server).
func (b *Branch) IsLocalOnly() bool {
	return "origin/"+b.Name != b.OriginBranch()
}

// HasPendingCommit reports whether b has any pending commits.
func (b *Branch) HasPendingCommit() bool {
	b.loadPending()
	return b.commitsAhead > 0
}

// Pending returns b's pending commits, newest first (children before parents).
func (b *Branch) Pending() []*Commit {
	b.loadPending()
	return b.pending
}

// Branchpoint returns an identifier for the latest revision
// common to both this branch and its upstream branch.
func (b *Branch) Branchpoint() string {
	b.loadPending()
	return b.branchpoint
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

func (b *Branch) loadPending() {
	if b.loadedPending {
		return
	}
	b.loadedPending = true

	// In case of early return.
	b.branchpoint = trim(cmdOutput("git", "rev-parse", HEAD))

	if b.DetachedHead() {
		return
	}

	// Note: --topo-order means child first, then parent.
	origin := b.OriginBranch()
	const numField = 5
	all := trim(cmdOutput("git", "log", "--topo-order", "--format=format:%H%x00%h%x00%P%x00%B%x00%s%x00", origin+".."+b.FullName(), "--"))
	fields := strings.Split(all, "\x00")
	if len(fields) < numField {
		return // nothing pending
	}
	for i, field := range fields {
		fields[i] = strings.TrimLeft(field, "\r\n")
	}
	foundMergeBranchpoint := false
	for i := 0; i+numField <= len(fields); i += numField {
		c := &Commit{
			Hash:      fields[i],
			ShortHash: fields[i+1],
			Parent:    strings.TrimSpace(fields[i+2]), // %P starts with \n for some reason
			Message:   fields[i+3],
			Subject:   fields[i+4],
		}
		if j := strings.Index(c.Parent, " "); j >= 0 {
			c.Parent, c.Merge = c.Parent[:j], c.Parent[j+1:]
			// Found merge point.
			// Merges break the invariant that the last shared commit (the branchpoint)
			// is the parent of the final commit in the log output.
			// If c.Parent is on the origin branch, then since we are reading the log
			// in (reverse) topological order, we know that c.Parent is the actual branchpoint,
			// even if we later see additional commits on a different branch leading down to
			// a lower location on the same origin branch.
			// Check c.Merge (the second parent) too, so we don't depend on the parent order.
			if strings.Contains(cmdOutput("git", "branch", "-a", "--contains", "--no-color", c.Parent), " "+origin+"\n") {
				foundMergeBranchpoint = true
				b.branchpoint = c.Parent
			}
			if strings.Contains(cmdOutput("git", "branch", "-a", "--contains", "--no-color", c.Merge), " "+origin+"\n") {
				foundMergeBranchpoint = true
				b.branchpoint = c.Merge
			}
		}
		for _, line := range lines(c.Message) {
			// Note: Keep going even if we find one, so that
			// we take the last Change-Id line, just in case
			// there is a commit message quoting another
			// commit message.
			// I'm not sure this can come up at all, but just in case.
			if strings.HasPrefix(line, "Change-Id: ") {
				c.ChangeID = line[len("Change-Id: "):]
			}
		}

		b.pending = append(b.pending, c)
		if !foundMergeBranchpoint {
			b.branchpoint = c.Parent
		}
	}
	b.commitsAhead = len(b.pending)
	b.commitsBehind = len(lines(cmdOutput("git", "log", "--format=format:x", b.FullName()+".."+b.OriginBranch(), "--")))
}

func branchPrefix(s string) (string, string) {
	for _, prefix := range dripBranchPrefixes() {
		if strings.HasPrefix(s, prefix) {
			return prefix, strings.TrimPrefix(s, prefix)
		}
	}
	return "", s
}

// LocalBranches returns a list of all known local branches.
// If the current directory is in detached HEAD mode, one returned
// branch will have Name == "HEAD" and DetachedHead() == true.
func LocalBranches() []*Branch {
	var branches []*Branch
	current := CurrentBranch()
	out, err := cmdOutputErr("git", "branch", "-q", "--no-color")
	if err != nil {
		return branches
	}
	for _, s := range nonBlankLines(out) {
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
				s = current.PrefixedName()
			} else {
				s = strings.TrimPrefix(s, "* ")
			}
		}
		prefix, name := branchPrefix(s)
		branches = append(branches, &Branch{Name: name, Prefix: prefix})
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
				s = current.PrefixedName()
			} else {
				s = strings.TrimPrefix(s, "* ")
			}
		}

		if strings.HasPrefix(s, prefix) {
			p, name := branchPrefix(s)
			branches = append(branches, &Branch{Name: name, Prefix: p})
		}
	}
	return branches
}

// OriginBranches returns a list of remote branches
func OriginBranches() []string {
	var branches []string
	for _, line := range nonBlankLines(cmdOutput("git", "branch", "-a", "-q")) {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, " -> "); i >= 0 {
			line = line[:i]
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "* "))
		if strings.HasPrefix(name, "remotes/origin/") {
			branches = append(branches, strings.TrimPrefix(name, "remotes/"))
		}
	}
	return branches
}

// IsRepoHeadless checks if the current git repo is headless
func IsRepoHeadless() bool {
	_, err := cmdOutputErr("git", "rev-parse", "--quiet", "--verify", "HEAD")
	return err != nil
}

func contains(branches []*Branch, name string) bool {
	for _, b := range branches {
		if b.PrefixedName() == name {
			return true
		}
	}
	return false
}

func remoteContains(name string) bool {
	for _, b := range OriginBranches() {
		if b == name {
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

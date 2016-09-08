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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var (
	runLog     []string
	testStderr *bytes.Buffer
	testStdout *bytes.Buffer
	died       bool
)

var gitversion = "unknown git version" // git version for error logs

type gitTest struct {
	pwd         string // current directory before test
	tmpdir      string // temporary directory holding repos
	server      string // server repo root
	client      string // client repo root
	nwork       int    // number of calls to work method
	nworkServer int    // number of calls to serverWork method
	nworkOther  int    // number of calls to serverWorkUnrelated method
}

// resetReadOnlyFlagAll resets windows read-only flag
// set on path and any children it contains.
// The flag is set by git and has to be removed.
// os.Remove refuses to remove files with read-only flag set.
func resetReadOnlyFlagAll(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return os.Chmod(path, 0666)
	}

	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = fd.Close()
	}()

	names, _ := fd.Readdirnames(-1)
	for _, name := range names {
		_ = resetReadOnlyFlagAll(path + string(filepath.Separator) + name)
	}
	return nil
}

func (gt *gitTest) done() {
	_ = os.Chdir(gt.pwd) // #nosec change out of gt.tmpdir first, otherwise following os.RemoveAll fails on windows
	_ = resetReadOnlyFlagAll(gt.tmpdir)
	_ = os.RemoveAll(gt.tmpdir)
}

// doWork simulates commit 'n' touching 'file' in 'dir'
func doWork(t *testing.T, n int, dir, file, changeid string) {
	write(t, dir+"/"+file, fmt.Sprintf("new content %d", n))
	trun(t, dir, "git", "add", file)
	suffix := ""
	if n > 1 {
		suffix = fmt.Sprintf(" #%d", n)
	}
	msg := fmt.Sprintf("msg%s\n\nChange-Id: I%d%s\n", suffix, n, changeid)
	trun(t, dir, "git", "commit", "-m", msg)
}

func (gt *gitTest) work(t *testing.T) {
	if gt.nwork == 0 {
		trun(t, gt.client, "git", "checkout", "-b", "work")
		trun(t, gt.client, "git", "branch", "--set-upstream-to", "origin/master")
		trun(t, gt.client, "git", "tag", "work") // make sure commands do the right thing when there is a tag of the same name
	}

	// make local change on client
	gt.nwork++
	doWork(t, gt.nwork, gt.client, "file", "23456789")
}

func (gt *gitTest) serverWork(t *testing.T) {
	// make change on server
	// duplicating the sequence of changes in gt.work to simulate them
	// having gone through Gerrit and submitted with possibly
	// different commit hashes but the same content.
	gt.nworkServer++
	doWork(t, gt.nworkServer, gt.server, "file", "23456789")
}

func (gt *gitTest) serverWorkUnrelated(t *testing.T) {
	// make unrelated change on server
	// this makes history different on client and server
	gt.nworkOther++
	doWork(t, gt.nworkOther, gt.server, "otherfile", "9999")
}

func gitCheck(t *testing.T) {
	// The Linux builders seem not to have git in their paths.
	// That makes this whole repo a bit useless on such systems,
	// but make sure the tests don't fail.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("cannot find git in path: %v", err)
	}
}

func newGitTestFolder(t *testing.T) (gt *gitTest) {
	gitCheck(t)

	tmpdir, err := ioutil.TempDir("", "git-drip-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if gt == nil {
			_ = os.RemoveAll(tmpdir) // #nosec
		}
	}()

	client := tmpdir + "/git-client"
	mkdir(t, client)

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(client); err != nil {
		t.Fatal(err)
	}

	return &gitTest{
		pwd:    pwd,
		tmpdir: tmpdir,
		server: "",
		client: client,
	}
}

func newGitTestInit(t *testing.T) (gt *gitTest) {
	gt = newGitTestFolder(t)

	trun(t, gt.client, "git", "init", ".")
	trun(t, gt.client, "git", "config", "user.name", "gopher")
	trun(t, gt.client, "git", "config", "user.email", "gopher@example.com")

	return
}

func newGitTest(t *testing.T) (gt *gitTest) {
	gitCheck(t)

	tmpdir, err := ioutil.TempDir("", "git-drip-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if gt == nil {
			_ = os.RemoveAll(tmpdir)
		}
	}()

	gitversion = trun(t, tmpdir, "git", "--version")

	server := tmpdir + "/git-origin"

	mkdir(t, server)
	write(t, server+"/file", "this is master")
	write(t, server+"/.gitattributes", "* -text\n")
	trun(t, server, "git", "init", ".")
	trun(t, server, "git", "config", "user.name", "gopher")
	trun(t, server, "git", "config", "user.email", "gopher@example.com")
	trun(t, server, "git", "add", "file", ".gitattributes")
	trun(t, server, "git", "commit", "-m", "on master")

	for _, name := range []string{"feature/test", "release/v1.0"} {
		trun(t, server, "git", "checkout", "master")
		trun(t, server, "git", "checkout", "-b", name)
		write(t, server+"/file."+name, "this is "+name)
		trun(t, server, "git", "add", "file."+name)
		trun(t, server, "git", "commit", "-m", "on "+name)
	}
	trun(t, server, "git", "checkout", "master")

	client := tmpdir + "/git-client"
	mkdir(t, client)
	trun(t, client, "git", "clone", server, ".")
	trun(t, client, "git", "config", "user.name", "gopher")
	trun(t, client, "git", "config", "user.email", "gopher@example.com")

	trun(t, client, "git", "config", "core.editor", "false")
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(client); err != nil {
		t.Fatal(err)
	}

	return &gitTest{
		pwd:    pwd,
		tmpdir: tmpdir,
		server: server,
		client: client,
	}
}

func mkdir(t *testing.T, dir string) {
	if err := os.Mkdir(dir, 0777); err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, file, data string) {
	if err := ioutil.WriteFile(file, []byte(data), 0666); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func remove(t *testing.T, file string) {
	if err := os.RemoveAll(file); err != nil {
		t.Fatal(err)
	}
}

func trun(t *testing.T, dir string, cmdline ...string) string {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if cmdline[0] == "git" {
			t.Fatalf("in %s/, ran %s with %s:\n%v\n%s", filepath.Base(dir), cmdline, gitversion, err, out)
		}
		t.Fatalf("in %s/, ran %s: %v\n%s", filepath.Base(dir), cmdline, err, out)
	}
	return string(out)
}

// fromSlash is like filepath.FromSlash, but it ignores ! at the start of the path
// and " (staged)" at the end.
func fromSlash(path string) string {
	if len(path) > 0 && path[0] == '!' {
		return "!" + fromSlash(path[1:])
	}
	if strings.HasSuffix(path, " (staged)") {
		return fromSlash(path[:len(path)-len(" (staged)")]) + " (staged)"
	}
	return filepath.FromSlash(path)
}

//########################################

func testRan(t *testing.T, cmds ...string) {
	if cmds == nil {
		cmds = []string{}
	}
	if !reflect.DeepEqual(runLog, cmds) {
		t.Errorf("ran:\n%s", strings.Join(runLog, "\n"))
		t.Errorf("wanted:\n%s", strings.Join(cmds, "\n"))
	}
}

func testPrinted(t *testing.T, buf *bytes.Buffer, name string, messages ...string) {
	all := buf.String()
	var errors bytes.Buffer
	for _, msg := range messages {
		if strings.HasPrefix(msg, "!") {
			if strings.Contains(all, msg[1:]) {
				fmt.Fprintf(&errors, "%s does (but should not) contain %q\n", name, msg[1:])
			}
			continue
		}
		if !strings.Contains(all, msg) {
			fmt.Fprintf(&errors, "%s does not contain %q\n", name, msg)
		}
	}
	if errors.Len() > 0 {
		t.Fatalf("wrong output\n%s%s:\n%s", &errors, name, all)
	}
}

func testPrintedStdout(t *testing.T, messages ...string) {
	testPrinted(t, testStdout, "stdout", messages...)
}

func testPrintedStderr(t *testing.T, messages ...string) {
	testPrinted(t, testStderr, "stderr", messages...)
}

func testNoStdout(t *testing.T) {
	if testStdout.Len() != 0 {
		t.Fatalf("unexpected stdout:\n%s", testStdout)
	}
}

func testNoStderr(t *testing.T) {
	if testStderr.Len() != 0 {
		t.Fatalf("unexpected stderr:\n%s", testStderr)
	}
}

var mainCanDie bool

func testMainDied(t *testing.T, args ...string) {
	mainCanDie = true
	testMain(t, args...)
	if !died {
		t.Fatalf("expected to die, did not\nstdout:\n%sstderr:\n%s", testStdout, testStderr)
	}
}

func testMainCanDie(t *testing.T, args ...string) {
	mainCanDie = true
	testMain(t, args...)
}

func testMain(t *testing.T, args ...string) {
	*noRun = false
	*verbose = 0

	t.Logf("git-drip %s", strings.Join(args, " "))

	canDie := mainCanDie
	mainCanDie = false // reset for next invocation

	defer func() {
		runLog = runLogTrap
		testStdout = stdoutTrap
		testStderr = stderrTrap

		dieTrap = nil
		runLogTrap = nil
		stdoutTrap = nil
		stderrTrap = nil
		if err := recover(); err != nil {
			if died && canDie {
				return
			}
			var msg string
			if died {
				msg = "died"
			} else {
				msg = fmt.Sprintf("panic: %v", err)
			}
			t.Fatalf("%s\nstdout:\n%sstderr:\n%s", msg, testStdout, testStderr)
		}
	}()

	dieTrap = func() {
		died = true
		panic("died")
	}
	died = false
	runLogTrap = []string{} // non-nil, to trigger saving of commands
	stdoutTrap = new(bytes.Buffer)
	stderrTrap = new(bytes.Buffer)

	os.Args = append([]string{"git-drip"}, args...)
	Execute()
}

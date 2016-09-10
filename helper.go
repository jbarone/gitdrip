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
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var verbose int
var noRun bool
var stdoutTrap, stderrTrap *bytes.Buffer

// SetVerboseLevel sets the level of verbosity for reporting
func SetVerboseLevel(level int) {
	verbose = level
}

// SetNoRun set whether or not to run the commands
func SetNoRun(b bool) {
	noRun = b
}

func stdout() io.Writer {
	if stdoutTrap != nil {
		return stdoutTrap
	}
	return os.Stdout
}

func stderr() io.Writer {
	if stderrTrap != nil {
		return stderrTrap
	}
	return os.Stderr
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(stderr(), format, args...)
	fmt.Fprintln(stderr(), "")
}

var dieTrap func()

func dief(format string, args ...interface{}) {
	printf(format, args...)
	die()
}

func die() {
	if dieTrap != nil {
		dieTrap()
	}
	os.Exit(1)
}

// cmdOutput runs the command line, returning its output.
// If the command cannot be run or does not exit successfully,
// cmdOutput dies.
//
// NOTE: cmdOutput must be used only to run commands that read state,
// not for commands that make changes. Commands that make changes
// should be run using runDirErr so that the -v and -n flags apply to them.
func cmdOutput(command string, args ...string) string {
	return cmdOutputDir(".", command, args...)
}

// cmdOutputDir runs the command line in dir, returning its output.
// If the command cannot be run or does not exit successfully,
// cmdOutput dies.
//
// NOTE: cmdOutput must be used only to run commands that read state,
// not for commands that make changes. Commands that make changes
// should be run using runDirErr so that the -v and -n flags apply to them.
func cmdOutputDir(dir, command string, args ...string) string {
	s, err := cmdOutputDirErr(dir, command, args...)
	if err != nil {
		fmt.Fprintf(stderr(), "%v\n%s\n", commandString(command, args), s)
		dief("%v", err)
	}
	return s
}

// cmdOutputErr runs the command line in dir, returning its output
// and any error results.
//
// NOTE: cmdOutputErr must be used only to run commands that read state,
// not for commands that make changes. Commands that make changes
// should be run using runDirErr so that the -v and -n flags apply to them.
func cmdOutputErr(command string, args ...string) (string, error) {
	return cmdOutputDirErr(".", command, args...)
}

// cmdOutputDirErr runs the command line in dir, returning its output
// and any error results.
//
// NOTE: cmdOutputDirErr must be used only to run commands that read state,
// not for commands that make changes. Commands that make changes
// should be run using runDirErr so that the -v and -n flags apply to them.
func cmdOutputDirErr(dir, command string, args ...string) (string, error) {
	// NOTE: We only show these non-state-modifying commands with -v -v.
	// Otherwise things like 'git sync -v' show all our internal "find out about
	// the git repo" commands, which is confusing if you are just trying to find
	// out what git sync means.
	if verbose > 1 {
		fmt.Fprintln(stderr(), commandString(command, args))
	}
	cmd := exec.Command(command, args...) // #nosec
	if dir != "." {
		cmd.Dir = dir
	}
	b, err := cmd.CombinedOutput()
	return string(b), err
}

func commandString(command string, args []string) string {
	return strings.Join(append([]string{command}, args...), " ")
}

func run(command string, args ...string) {
	if err := runErr(command, args...); err != nil {
		if verbose == 0 {
			// If we're not in verbose mode, print the command
			// before dying to give context to the failure.
			fmt.Fprintf(stderr(), "(running: %s)\n",
				commandString(command, args))
		}
		dief("%v", err)
	}
}

func runErr(command string, args ...string) error {
	return runDirErr("", command, args...)
}

var runLogTrap []string

func runDirErr(dir, command string, args ...string) error {
	if verbose > 0 || noRun {
		fmt.Fprintln(stderr(), commandString(command, args))
	}
	if noRun {
		return nil
	}
	if runLogTrap != nil {
		runLogTrap = append(runLogTrap,
			strings.TrimSpace(command+" "+strings.Join(args, " ")))
	}
	cmd := exec.Command(command, args...) // #nosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout()
	cmd.Stderr = stderr()
	return cmd.Run()
}

// trim is shorthand for strings.TrimSpace.
func trim(text string) string {
	return strings.TrimSpace(text)
}

// trimErr applies strings.TrimSpace to the result of cmdOutput(Dir)Err,
// passing the error along unmodified.
func trimErr(text string, err error) (string, error) {
	return strings.TrimSpace(text), err
}

// lines returns the lines in text.
func lines(text string) []string {
	out := strings.Split(text, "\n")
	// Split will include a "" after the last line. Remove it.
	if n := len(out) - 1; n >= 0 && out[n] == "" {
		out = out[:n]
	}
	return out
}

// nonBlankLines returns the non-blank lines in text.
func nonBlankLines(text string) []string {
	var out []string
	for _, s := range lines(text) {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

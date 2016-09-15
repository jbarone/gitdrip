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
	"testing"
)

func testVersion(t *testing.T) {
	defer func() {
		runLog = runLogTrap
		testStdout = stdoutTrap
		testStderr = stderrTrap

		dieTrap = nil
		runLogTrap = nil
		stdoutTrap = nil
		stderrTrap = nil
		if err := recover(); err != nil {
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

	PrintVersion()
}

func TestPrintVersion(t *testing.T) {
	testVersion(t)
	testPrintedStdout(t, "0.3.2")
}

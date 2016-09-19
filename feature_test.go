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

import "testing"

func testFeature(t *testing.T, canDie, descriptions bool) {
	defer testSetup(t)(canDie)

	RequireDripInitialized()
	ListFeatures(descriptions)
}

func TestFeatureInitRequire(t *testing.T) {
	gt := newGitTest(t)
	defer gt.done()

	testFeature(t, true, false)
	testPrintedStderr(t, "fatal: Not a git-drip enabled repo yet. "+
		"Please run \"git drip init\" first.\n")
}

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

const (
	dripMaster  = "gitdrip.branch.master"
	dripFeature = "gitdrip.prefix.feature"
	dripRelease = "gitdrip.prefix.release"
	dripHotfix  = "gitdrip.prefix.hotfix"
	dripVersion = "gitdrip.prefix.versiontag"
	dripOrigin  = "gitdrip.origin"
)

var prefixes []string

// DripBranchPrefixes returns a list of prefixes used for branches
func DripBranchPrefixes() []string {
	if len(prefixes) == 0 {
		for _, prefix := range []string{
			dripFeature,
			dripRelease,
			dripHotfix,
		} {
			if p := Config().Get(prefix); p != "" {
				prefixes = append(prefixes, p)
			}
		}
	}
	return prefixes
}

func isDripMasterConfigured() bool {
	if !Config().Has(dripMaster) || Config().Get(dripMaster) == "" ||
		!contains(LocalBranches(), Config().Get(dripMaster)) {
		return false
	}
	return true
}

func areDripPrefixConfigured() bool {
	for _, prefix := range []string{
		dripFeature,
		dripRelease,
		dripHotfix,
		dripVersion,
	} {
		if !Config().Has(prefix) {
			return false
		}
	}

	return true
}

// DripInitialized returns whether or not the git repository is configured
// for the git-drip workflow.
func DripInitialized() bool {
	return isDripMasterConfigured() && areDripPrefixConfigured()
}

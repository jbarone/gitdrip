package gitdrip

import (
	"fmt"
	"strings"
)

func requireCleanRepo() {
	if GitDir() == "" {
		run("git", "init")
	}

	if !IsRepoHeadless() {
		requireCleanTree()
	}
}

func getDevelopSuggestion(force bool) (string, bool) {
	var suggestion string
	var shouldCheck bool
	var branches = LocalBranches()

	if len(branches) == 0 { // fresh repo
		fmt.Fprintln(stdout(),
			"No branches exist yet. Base branches must be created now.")
		if suggestion = Config().Get(dripMaster); suggestion == "" {
			suggestion = "master"
		}
	} else {
		fmt.Fprintln(stdout(),
			"\nWhich branch should be used for development?")
		for _, b := range branches {
			fmt.Fprintf(stdout(), "   - %s\n", b.Name)
		}
		shouldCheck = true
		suggestion = Config().Get(dripMaster)
		if suggestion == "" || !branchesContains(branches, suggestion) {
			suggestion = "master"
		}
	}
	return suggestion, shouldCheck
}

func configureDevelop(force, defaults bool) string {
	if isDripMasterConfigured() && !force {
		return Config().Get(dripMaster)
	}

	var answer, masterBranch string

	suggestion, shouldCheck := getDevelopSuggestion(force)

	fmt.Fprintf(stdout(), "Branch name for development: [%s] ", suggestion)
	if defaults {
		fmt.Fprintln(stdout(), "")
		masterBranch = suggestion
	} else {
		fmt.Scanln(&answer)
		if strings.TrimSpace(answer) == "" {
			masterBranch = suggestion
		} else {
			masterBranch = answer
		}
	}

	// check existence in case of an already existing repo
	if shouldCheck && !branchesContains(LocalBranches(), masterBranch) {
		dief("Local branch '%s' does not exist.", masterBranch)
	}

	Config().Set(dripMaster, masterBranch)

	return masterBranch
}

func enforceHead(masterBranch string) {
	if IsRepoHeadless() {
		run("git", "symbolic-ref", "HEAD",
			fmt.Sprintf("refs/heads/%s", masterBranch))
		run("git", "commit", "--allow-empty", "--quiet",
			"-m", "Initial commit")
		run("git", "checkout", "-q", masterBranch)
	}
}

func configurePrefixes(force, defaults bool) {
	config := Config()
	if force || !config.Has(dripFeature) ||
		!config.Has(dripRelease) ||
		!config.Has(dripHotfix) ||
		!config.Has(dripVersion) {
		var suggestion, prefix, answer string

		fmt.Fprintln(stdout(), "\nHow to name supporting branch prefixes?")

		for _, ps := range []struct {
			key        string
			suggestion string
			question   string
		}{
			{
				dripFeature,
				"feature/",
				"Feature branches? [%s] ",
			},
			{
				dripRelease,
				"release/",
				"Release branches? [%s] ",
			},
			{
				dripHotfix,
				"hotfix/",
				"Hotfix branches? [%s] ",
			},
			{
				dripVersion,
				"",
				"Version tag prefix? [%s] ",
			},
		} {

			if suggestion = config.Get(ps.key); suggestion == "" {
				suggestion = ps.suggestion
			}
			fmt.Fprintf(stdout(), ps.question, suggestion)
			if defaults {
				fmt.Fprintln(stdout(), "")
				prefix = suggestion
			} else {
				fmt.Scanln(&answer)
				if strings.TrimSpace(answer) == "" {
					prefix = suggestion
				} else {
					prefix = answer
				}
			}

			run("git", "config", ps.key, prefix)
		}
	}
}

// InitDrip initializes a git-drip repo
func InitDrip(force, defaults bool) {
	requireCleanRepo()

	if DripInitialized() && !force {
		dief(`Already initialized for git-drip.
To force reinitialization, use: git drip init -f`)
	}

	if defaults {
		fmt.Fprintln(stdout(), "Using default branch names.")
	}

	var masterBranch = configureDevelop(force, defaults)
	enforceHead(masterBranch)

	configurePrefixes(force, defaults)

	fmt.Fprintln(stdout(), "\ngit drip has been initialized")
}

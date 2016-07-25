# git-drip

A collection of Git extensions to provide high-level repository operations for
the git-drip workflow

## Installing git-drip

Clone the git-drip sources from Github:

    $ git clone git://github.com/jbarone/gitdrip.git

Install using Make:

    $ sudo make install

By default, git-drip will be installed int /usr/local. To change the prefix
where git-drip will be installed, simply specify it explicitly, using:

    $ sudo make prefix=/opt/local install

### Initialization

To initialize a new repo with the basic branch structure, use:

    $ git drip init

This will then interactively prompt you with some questions on which branches
you would like to use as development branch, and how you
would like your prefixes to be named. You may simply press Return on any of
those questions to accept the (sane) default suggestions.

### Creating feature/release/hotfix branches

* To list/start/finish feature branches, use:

        $ git drip feature
        $ git drip feature start <name> [<base>]
        $ git drip feature finish <name>

* To list/start/finish release branches, use:

        $ git drip release
        $ git drip release start <release> [<base>]
        $ git drip release finish <release>

* To list/start/finish hotfix branches, use:

        $ git drip hotfix
        $ git drip hotfix start <release> <base>
        $ git drip hotfix finish <release>

  For hotfix branches, the `<base>` argument must be a commit on `master`.

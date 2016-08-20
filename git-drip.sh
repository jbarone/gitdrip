#!/bin/sh
# vim:et:ft=sh:sts=4:sw=4

# enable debug mode
if [ "$DEBUG" = "yes"  ]; then
    set -x
fi

export GITDRIP_DIR=$(dirname "$0")

usage() {
    echo "usage: git drip <subcommand>"
    echo
    echo "Available subcommands are:"
    echo "   init      Initialize a new git repo with support for the workflow."
    echo "   feature   Manage your feature branches."
    echo "   release   Manage your release branches."
    echo "   hotfix    Manage your hotfix branches."
    echo "   version   Shows version information."
    echo
    echo "Try 'git drip <subcommand> help' for details."
}

main() {
    if [ $# -lt 1 ]; then
        usage
        exit 1
    fi

    # load common functionality
    . "$GITDRIP_DIR/gitdrip-common"

    # This environmental variable fixes non-POSIX getopt style argument
    # parsing, effectively breaking git-drip subcommand parsing on several
    # Linux platforms.
    export POSIXLY_CORRECT=1

    # use the shFlags project to parse the command line arguments
    . "$GITDRIP_DIR/gitdrip-shFlags"
    FLAGS_PARENT="git drip"

    # do actual parsing
    FLAGS "$@" || exit $?
    eval set -- "${FLAGS_ARGV}"

    # sanity checks
    SUBCOMMAND="$1"; shift

    if [ ! -e "$GITDRIP_DIR/git-drip-$SUBCOMMAND" ]; then
        usage
        exit 1
    fi

    # run command
    . "$GITDRIP_DIR/git-drip-$SUBCOMMAND"
    FLAGS_PARENT="git drip $SUBCOMMAND"

    # test if the first argument is a flag (i.e. starts with '-')
    # in that case, we interpret this arg as a flag for the default
    # command
    SUBACTION="default"
    if [ "$1" != "" ] && { ! echo "$1" | grep -q "^-"; } then
        SUBACTION="$1"; shift
    fi
    if ! type "cmd_$SUBACTION" >/dev/null 2>&1; then
        warn "Unknown subcommand: '$SUBACTION'"
        usage
        exit 1
    fi

    # run the specified action
    cmd_$SUBACTION "$@"
}

main "$@"

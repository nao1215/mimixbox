# Helpers exercising the top-level MimixBox CLI surface (GitHub issue #234):
# help/list output, unknown-option handling, and an install/remove smoke path
# against a temporary directory.

MbHelp() {
    mimixbox --help
}

MbList() {
    mimixbox --list
}

MbUnknownOption() {
    mimixbox --definitely-not-an-option
}

# MbInstallRemoveSmoke creates applet symlinks in a fresh temp directory, then
# removes them, asserting the round trip end to end. It prints "ok" only when
# every step behaves as expected so the spec can assert on a single token.
MbInstallRemoveSmoke() {
    d=$(mktemp -d) || return 1
    # shellcheck disable=SC2064
    trap "rm -rf '$d'" EXIT

    mimixbox --full-install "$d" >/dev/null 2>&1 || return 1
    [ -L "$d/cat" ] || return 1     # a representative applet symlink exists
    [ -L "$d/pidof" ] || return 1

    mimixbox --remove "$d" >/dev/null 2>&1 || return 1
    [ -L "$d/cat" ] && return 1     # MimixBox-owned symlinks are gone
    [ -L "$d/pidof" ] && return 1

    echo ok
}

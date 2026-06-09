# Helpers asserting that the end-to-end harness is hermetic: every applet on
# PATH must resolve to the MimixBox binary, never to a host command of the same
# name (GitHub issue #270).

# ResolvesToMimixBox prints "mimixbox" when the named command resolves, through
# the MimixBox-installed symlink, to the MimixBox binary. It prints whatever the
# real target's basename is otherwise, so a host binary makes the test fail
# loudly instead of passing silently.
ResolvesToMimixBox() {
    path=$(command -v "$1") || return 1
    # Follow the symlink chain to the concrete binary it points at.
    real=$(readlink -f "${path}")
    basename "${real}"
}

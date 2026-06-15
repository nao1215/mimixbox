# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "sha384sum" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'sha384sum'
  It 'describes itself with --help'
    When run command sha384sum --help
    The status should be success
    The output should include 'Usage: sha384sum'
    The stderr should equal ''
  End
  It 'prints the SHA-384 digest of stdin'
    Data 'hello'
    When run command sha384sum
    The status should be success
    The output should equal '1d0f284efe3edea4b9ca3bd514fa134b17eae361ccc7a1eefeff801b9bd6604e01f21f6bf249ef030599f0c218f2ba8c  -'
  End
End

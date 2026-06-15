# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "uuencode" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'uuencode'
  It 'describes itself with --help'
    When run command uuencode --help
    The status should be success
    The output should include 'Usage: uuencode'
    The stderr should equal ''
  End
  It 'uuencodes stdin with a begin header'
    Data 'hello'
    When run command uuencode hi.txt
    The status should be success
    The output should include 'begin 644 hi.txt'
    The output should include 'end'
  End
End

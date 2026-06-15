# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "sum" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'sum'
  It 'describes itself with --help'
    When run command sum --help
    The status should be success
    The output should include 'Usage: sum'
    The stderr should equal ''
  End
  It 'documents its purpose in --help'
    When run command sum --help
    The status should be success
    The output should include 'BSD algorithm'
  End
  It 'prints a BSD checksum and block count for stdin'
    Data 'hello'
    When run command sum
    The status should be success
    The output should equal '36979     1'
  End
End

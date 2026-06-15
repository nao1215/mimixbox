# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "groups" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'groups'
  It 'describes itself with --help'
    When run command groups --help
    The status should be success
    The output should include 'Usage: groups'
    The stderr should equal ''
  End
  It 'prints the groups of a named user'
    When run command groups root
    The status should be success
    The output should equal 'root'
  End
End

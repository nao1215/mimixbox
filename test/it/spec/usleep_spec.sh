# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "usleep" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'usleep'
  It 'describes itself with --help'
    When run command usleep --help
    The status should be success
    The output should include 'Usage: usleep'
    The stderr should equal ''
  End
  It 'rejects a non-numeric microsecond count'
    When run command usleep notanumber
    The status should be failure
    The stderr should include 'usleep:'
  End
End

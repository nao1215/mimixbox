# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "http-status-code" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'http-status-code'
  It 'describes itself with --help'
    When run command http-status-code --help
    The status should be success
    The output should include 'Usage: http-status-code'
    The stderr should equal ''
  End
  It 'looks up a status code by number'
    When run command http-status-code search 404
    The status should be success
    The output should include '404 Not Found'
  End
End

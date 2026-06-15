# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "ipcalc" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'ipcalc'
  It 'describes itself with --help'
    When run command ipcalc --help
    The status should be success
    The output should include 'Usage: ipcalc'
    The stderr should equal ''
  End
  It 'documents its purpose in --help'
    When run command ipcalc --help
    The status should be success
    The output should include 'IPv4 network parameters'
  End
End

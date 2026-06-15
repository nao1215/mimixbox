# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "factor" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'factor'
  It 'describes itself with --help'
    When run command factor --help
    The status should be success
    The output should include 'Usage: factor'
    The stderr should equal ''
  End
  It 'documents its purpose in --help'
    When run command factor --help
    The status should be success
    The output should include 'Print the prime factors'
  End
  It 'factors a small integer'
    When run command factor 12
    The status should be success
    The output should equal '12: 2 2 3'
  End
  It 'fails on a non-numeric operand'
    When run command factor notanumber
    The status should be failure
    The stderr should include 'factor:'
  End
End

# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "egrep" applet.
# egrep is grep -E; it renders its own egrep-named structured --help. This file is the dedicated
# spec required by #477. Shared matching behavior is covered by the grep family.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'egrep'
  It 'describes itself with --help'
    When run command egrep --help
    The status should be success
    The output should include 'Usage: egrep'
    The stderr should equal ''
  End
End

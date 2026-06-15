# shellcheck shell=sh
# Issue #477: dedicated shell-level contract spec for the "crc32" applet.
#
# Every MimixBox applet's --help is rendered by internal/command's
# FlagSet.WriteUsage, so it exits 0, prints a "Usage: <cmd>" line, and writes
# nothing to stderr. That universal contract is asserted here; privileged,
# networked, and destructive applets are exercised via --help only so the
# suite never reboots, formats, loads modules, or touches the network.
Describe 'crc32'
  It 'describes itself with --help'
    When run command crc32 --help
    The status should be success
    The output should include 'Usage: crc32'
    The stderr should equal ''
  End
  It 'prints the CRC-32 of stdin'
    Data 'hello'
    When run command crc32
    The status should be success
    The output should equal '363a3020  -'
  End
End

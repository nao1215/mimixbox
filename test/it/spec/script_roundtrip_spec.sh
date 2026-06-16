# shellcheck shell=sh
# Issue #787: end-to-end round-trip for the "script" and "scriptreplay" applets.
#
# script -c COMMAND -T TIMINGFILE TYPESCRIPT records COMMAND's output, wrapping
# it in "Script started"/"Script done" framing, and writes a "delay bytes"
# timing file. scriptreplay TIMINGFILE TYPESCRIPT then re-emits the captured
# payload. This spec records a deterministic two-line payload, asserts the
# transcript framing + payload and the timing-file record shape, then replays
# it and asserts the payload bytes survive the round trip.
Describe 'script / scriptreplay round-trip'
  setup() {
    work="$(it_root)/script_roundtrip"
    rm -rf "$work"
    mkdir -p "$work"
    transcript="$work/transcript"
    timing="$work/timing"
    script -c 'printf "hello\nworld\n"' -T "$timing" "$transcript" >/dev/null 2>&1
  }
  cleanup() { rm -rf "$work"; }
  BeforeEach 'setup'
  AfterEach 'cleanup'

  It 'records the transcript framing'
    When run cat "$transcript"
    The status should be success
    The output should include 'Script started'
    The output should include 'Script done'
  End

  It 'records the command payload in the transcript'
    When run cat "$transcript"
    The status should be success
    The output should include 'hello'
    The output should include 'world'
  End

  It 'writes a timing file of "delay bytes" records'
    # Every line is two whitespace-separated fields: a float delay and an
    # integer byte count.
    When run grep -Eq '^[0-9]+\.[0-9]+[[:space:]]+[0-9]+$' "$timing"
    The status should be success
  End

  It 'replays the captured payload from the timing + transcript'
    When run scriptreplay "$timing" "$transcript"
    The status should be success
    The output should include 'hello'
    The output should include 'world'
  End
End

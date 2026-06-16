# shellcheck shell=sh
# Issue #789: real-process alias-parity contract for the thin wrapper applets.
#
# These specs run the installed applets on PATH so the parity holds for the
# shipped binaries, not just the in-memory dispatch covered by the Go unit
# tests:
#   - egrep is grep -E and fgrep is grep -F, so the alias and grep with the
#     forced mode flag must produce the same match output over the same file.
#   - netcat is an alias of nc; both must answer --help with exit 0 and an
#     Examples section. Help short-circuits before any network code, so no
#     sockets are opened.
Describe 'alias parity'
  setup() {
    TEST_DIR=${MIMIXBOX_IT_ROOT}/alias_parity
    mkdir -p "${TEST_DIR}"
    # A regex metacharacter in the data makes -E (regex) and -F (fixed) modes
    # observably distinct, so the egrep/fgrep parity below is meaningful.
    printf 'apple\nbanana\na.b\naxb\n' > "${TEST_DIR}/fixture.txt"
  }
  cleanup() { rm -rf "${MIMIXBOX_IT_ROOT}/alias_parity"; }
  BeforeEach 'setup'
  AfterEach 'cleanup'

  fixture() { printf '%s' "${MIMIXBOX_IT_ROOT}/alias_parity/fixture.txt"; }

  Describe 'egrep == grep -E'
    It 'matches the same lines as grep -E over the same file'
      When run command egrep 'a(p|x)' "$(fixture)"
      The status should be success
      The output should equal "$(grep -E 'a(p|x)' "$(fixture)")"
    End
  End

  Describe 'fgrep == grep -F'
    It 'matches the same lines as grep -F over the same file'
      # The fixed-string pattern "a.b" must match only the literal a.b line, not
      # axb, exactly as grep -F does.
      When run command fgrep 'a.b' "$(fixture)"
      The status should be success
      The output should equal "$(grep -F 'a.b' "$(fixture)")"
    End
  End

  Describe 'netcat is an alias of nc'
    It 'answers netcat --help with exit 0 and an Examples section'
      When run command netcat --help
      The status should be success
      The output should include 'Usage: netcat'
      The output should include 'Examples:'
      The stderr should equal ''
    End

    It 'answers nc --help with exit 0 and an Examples section'
      When run command nc --help
      The status should be success
      The output should include 'Usage: nc'
      The output should include 'Examples:'
      The stderr should equal ''
    End
  End
End

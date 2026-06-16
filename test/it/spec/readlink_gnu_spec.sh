# shellcheck shell=sh
# Integration tests for the GNU flags added to readlink (issue #734):
#   --canonicalize-existing/-e, --canonicalize-missing/-m, --zero/-z.

Describe 'readlink GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/readlink_gnu"
        rm -rf "$WORK"
        mkdir -p "$WORK"
        : > "$WORK/target"
        ln -s "$WORK/target" "$WORK/link"
    }
    cleanup() {
        rm -rf "$WORK"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'fails when -e is given a missing path'
        When run readlink -e "$WORK/does-not-exist"
        The status should be failure
        The stdout should equal ''
    End

    It 'succeeds when -e is given an existing symlink'
        When run readlink -e "$WORK/link"
        The status should be success
        The stdout should equal "$WORK/target"
    End

    It 'succeeds with -m on a missing path'
        When run readlink -m "$WORK/a/b/c"
        The status should be success
        The stdout should equal "$WORK/a/b/c"
    End

    It 'terminates output with NUL under -z'
        check() {
            # Capture the raw output and confirm it ends with a NUL byte and not
            # a newline. od shows the trailing byte sequence.
            readlink -z "$WORK/link" | od -An -c | tr -s ' ' | grep -q '\\0$' && printf 'nul'
        }
        When call check
        The output should equal 'nul'
        The status should be success
    End
End

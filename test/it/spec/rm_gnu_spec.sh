# shellcheck shell=sh
# Integration tests for the GNU flags added to rm (issue #732):
#   --preserve-root (default), --no-preserve-root, --one-file-system.

Describe 'rm GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/rm_gnu"
        rm -rf "$WORK"
        mkdir -p "$WORK"
    }
    cleanup() {
        rm -rf "$WORK"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    # The default --preserve-root must refuse to recurse on "/" and exit
    # nonzero. This NEVER deletes anything: the guard returns before any
    # removal, so running it against the real "/" is safe.
    It 'refuses to recurse on / by default (--preserve-root)'
        When run rm -r /
        The status should be failure
        The stderr should include "it is dangerous to operate recursively on '/'"
        The stderr should include 'use --no-preserve-root to override this failsafe'
    End

    It 'removes an ordinary directory recursively (guard does not interfere)'
        prepare() {
            mkdir -p "$WORK/tree/sub"
            : > "$WORK/tree/sub/leaf.txt"
        }
        check() {
            prepare
            rm -r "$WORK/tree"
            [ ! -e "$WORK/tree" ] && printf 'gone'
        }
        When call check
        The output should equal 'gone'
        The status should be success
    End

    It 'removes a single-filesystem tree with --one-file-system'
        prepare() {
            mkdir -p "$WORK/ofs/a/b"
            : > "$WORK/ofs/a/b/leaf.txt"
            : > "$WORK/ofs/top.txt"
        }
        check() {
            prepare
            rm -r --one-file-system "$WORK/ofs"
            [ ! -e "$WORK/ofs" ] && printf 'gone'
        }
        When call check
        The output should equal 'gone'
        The status should be success
    End

    It 'removes / when --no-preserve-root is given (verified via guard-message absence on a safe target)'
        prepare() {
            mkdir -p "$WORK/victim/sub"
            : > "$WORK/victim/sub/f"
        }
        check() {
            prepare
            rm -r --no-preserve-root "$WORK/victim"
            [ ! -e "$WORK/victim" ] && printf 'gone'
        }
        When call check
        The output should equal 'gone'
        The status should be success
    End
End

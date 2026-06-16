# shellcheck shell=sh
# Integration tests for the GNU flags added to touch (issue #733):
#   --reference/-r, --date/-d, --time=WORD, --no-dereference/-h.

Describe 'touch GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/touch_gnu"
        rm -rf "$WORK"
        mkdir -p "$WORK"
    }
    cleanup() {
        rm -rf "$WORK"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    # SYSSTAT is the system stat(1), addressed by absolute path so the mimixbox
    # 'stat' applet on PATH does not shadow it. %Y is the mtime in epoch seconds.
    SYSSTAT=/usr/bin/stat

    It 'copies the reference file mtime (--reference)'
        check() {
            : > "$WORK/ref"
            : > "$WORK/dst"
            # Give ref a known mtime, then copy it onto dst.
            touch -d '2001-02-03 04:05:06' "$WORK/ref"
            touch --reference="$WORK/ref" "$WORK/dst"
            r=$("$SYSSTAT" -c '%Y' "$WORK/ref")
            d=$("$SYSSTAT" -c '%Y' "$WORK/dst")
            [ "$r" = "$d" ] && printf 'match'
        }
        When call check
        The output should equal 'match'
        The status should be success
    End

    It 'sets a known time with --date'
        check() {
            : > "$WORK/f"
            touch -d '2020-06-15 12:34:56' "$WORK/f"
            # Expected epoch for that local time, computed by the system date.
            want=$(/bin/date -d '2020-06-15 12:34:56' '+%s')
            got=$("$SYSSTAT" -c '%Y' "$WORK/f")
            [ "$want" = "$got" ] && printf 'match'
        }
        When call check
        The output should equal 'match'
        The status should be success
    End

    It 'accepts --time=atime without error'
        When run touch --time=atime "$WORK/g"
        The status should be success
        The path "$WORK/g" should be exist
    End

    It 'rejects an invalid --time word'
        When run touch --time=bogus "$WORK/h"
        The status should be failure
        The stderr should include 'invalid argument'
    End

    It 'changes the symlink itself with --no-dereference (-h)'
        check() {
            : > "$WORK/target"
            touch -d '2005-05-05 05:05:05' "$WORK/target"
            ln -s "$WORK/target" "$WORK/link"
            # -h must touch the link, leaving the target's mtime untouched.
            touch -h -d '2030-01-01 00:00:00' "$WORK/link"
            want=$(/bin/date -d '2005-05-05 05:05:05' '+%s')
            got=$("$SYSSTAT" -c '%Y' "$WORK/target")
            [ "$want" = "$got" ] && printf 'target-unchanged'
        }
        When call check
        The output should equal 'target-unchanged'
        The status should be success
    End
End

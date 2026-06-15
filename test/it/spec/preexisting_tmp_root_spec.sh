# Regression for issue #478: the integration suite must use a per-run temp root
# (MIMIXBOX_IT_ROOT) and must NOT assume /tmp/mimixbox is a writable directory.
#
# Historically the suite created children under a global /tmp/mimixbox/it, which
# broke `make it` (e.g. "mkdir /tmp/mimixbox: not a directory") whenever
# /tmp/mimixbox already existed as a regular file. Here we deliberately make
# /tmp/mimixbox a regular file (the worst case for the old assumption) and assert
# that the suite's per-run root is allocated elsewhere and is usable.

Describe 'issue #478: per-run temp root survives a pre-existing /tmp/mimixbox file'
    LEGACY_ROOT=/tmp/mimixbox
    CREATED_LEGACY=no

    setup_legacy() {
        # Only create the fixture if nothing is there, so we never clobber a
        # developer's real /tmp/mimixbox (file or directory).
        if [ ! -e "$LEGACY_ROOT" ]; then
            : > "$LEGACY_ROOT"
            CREATED_LEGACY=yes
        fi
    }

    cleanup_legacy() {
        # Remove only what we created, and only if it is still the regular file
        # we made.
        if [ "$CREATED_LEGACY" = yes ] && [ -f "$LEGACY_ROOT" ]; then
            rm -f "$LEGACY_ROOT"
        fi
    }

    BeforeAll 'setup_legacy'
    AfterAll  'cleanup_legacy'

    It 'allocates a usable per-run root that is not /tmp/mimixbox'
        check_root() {
            [ -n "${MIMIXBOX_IT_ROOT:-}" ] || return 1
            [ "${MIMIXBOX_IT_ROOT}" != "/tmp/mimixbox" ] || return 1
            mkdir -p "${MIMIXBOX_IT_ROOT}/it478" || return 1
            touch "${MIMIXBOX_IT_ROOT}/it478/probe" || return 1
            printf 'ok'
        }
        When call check_root
        The output should equal 'ok'
        The status should be success
    End

    It 'leaves a pre-existing /tmp/mimixbox file untouched (never turns it into a dir)'
        check_legacy() {
            # If we did not own the fixture, do not assert against the dev's state.
            if [ "$CREATED_LEGACY" != yes ]; then
                printf 'not-a-dir'
                return 0
            fi
            [ ! -d /tmp/mimixbox ] && printf 'not-a-dir'
        }
        When call check_legacy
        The output should equal 'not-a-dir'
        The status should be success
    End
End

Describe 'realpath GNU relative/logical flags (issue #757)'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/realpath_gnu"
        mkdir -p "$WORK/a/b"
    }
    cleanup() { rm -rf "${MIMIXBOX_IT_ROOT}/realpath_gnu"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'prints a path relative to --relative-to'
        When run realpath --relative-to "$WORK/a" "$WORK/a/b"
        The status should be success
        The output should equal 'b'
    End

    It 'resolves .. lexically with -L -m'
        When run realpath -L -m /tmp/../etc/./hosts
        The status should be success
        The output should equal '/etc/hosts'
    End
End

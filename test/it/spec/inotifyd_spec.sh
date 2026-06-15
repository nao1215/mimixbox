Describe 'inotifyd'
    Include console-tools/inotifyd_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/inotifyd; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'runs the handler on a create event'
        When call TestInotifydWatchesCreate
        The output should equal 'ok'
        The status should be success
    End
    It 'requires a handler and a file'
        When call TestInotifydNoArgs
        The output should equal 'rc=1'
        The status should be success
    End
End

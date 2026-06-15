Describe 'start-stop-daemon'
    Include loginutils/start_stop_daemon_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/ssd; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'starts and stops a background program'
        When call TestSsdStartStop
        The output should equal 'stopped'
        The status should be success
    End
    It 'requires a start or stop mode'
        When call TestSsdNoMode
        The output should equal 'rc=1'
        The status should be success
    End
End

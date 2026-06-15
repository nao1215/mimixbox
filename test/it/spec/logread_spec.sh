Describe 'logread'
    Include procps/logread_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/logread; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'prints a given log file'
        When call TestLogreadFile
        The line 1 of output should equal 'msg one'
        The line 2 of output should equal 'msg two'
        The status should be success
    End
    It 'fails when no readable log is found'
        When call TestLogreadMissing
        The output should equal 'rc=1'
        The status should be success
    End
End

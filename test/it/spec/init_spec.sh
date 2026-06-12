Describe 'init'
    Include loginutils/init_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/init; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'runs the inittab sysinit and wait actions'
        When call TestInitRunsActions
        The line 1 of output should equal 'SYSINIT'
        The line 2 of output should equal 'WAIT'
        The status should be success
    End
    It 'fails on a missing inittab'
        When call TestInitMissing
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'flock'
    Include util-linux/flock_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/flock; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'runs a command while holding the lock'
        When call TestFlockRuns
        The output should equal 'locked-run'
        The status should be success
    End
    It 'fails -n when the lock is already held'
        When call TestFlockNonblock
        The output should equal 'rc=1'
        The status should be success
    End
End

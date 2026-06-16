Describe 'tail --follow --pid'
    Include textutils/tail_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'stops following once the watched process exits'
        When call TestTailFollowPid
        The output should include 'start'
        The output should include 'appended'
        # A success status means tail exited on its own after the PID died,
        # rather than being killed by timeout (which would yield 124).
        The status should be success
    End
End

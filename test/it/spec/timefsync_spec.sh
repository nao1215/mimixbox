Describe 'time / fsync'
    Include shellutils/timefsync_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'time runs the command and passes its output through'
        When call TestTimeOutput
        The output should equal 'timed'
        The status should be success
    End
    It 'time reports the real elapsed time on stderr'
        When call TestTimeReportsReal
        The output should equal '1'
        The status should be success
    End
    It 'fsync succeeds on an existing file'
        When call TestFsync
        The output should equal '0'
        The status should be success
    End
    It 'fsync fails on a missing file'
        When call TestFsyncMissing
        The output should equal '1'
        The status should be success
    End
End

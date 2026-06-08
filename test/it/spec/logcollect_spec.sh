Describe 'log-collect'
    Include shellutils/logcollect_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'copies log files into the output directory'
        When call TestLogCollect
        The output should equal 'log'
        The status should be success
    End
End

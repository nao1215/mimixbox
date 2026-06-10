Describe 'logger'
    Include procps/logger_test.sh

    It 'rejects an unknown facility'
        When call TestLoggerBadPriority
        The output should equal 'rc=1'
        The status should be success
    End
End

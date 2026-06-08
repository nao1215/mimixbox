Describe 'logname'
    Include shellutils/logname_test.sh
    It 'prints the login name from LOGNAME'
        When call TestLogname
        The output should equal 'mimixuser'
        The status should be success
    End
End

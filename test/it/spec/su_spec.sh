Describe 'su'
    Include loginutils/su_test.sh

    It 'fails for an unknown user'
        When call TestSuUnknownUser
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestSuHelp
        The status should be success
        The output should include 'Usage: su'
        The output should include 'user'
    End
End

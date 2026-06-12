Describe 'login'
    Include loginutils/login_test.sh

    It 'fails for an unknown user'
        When call TestLoginUnknownUser
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestLoginHelp
        The status should be success
        The output should include 'Usage: login'
        The output should include 'login shell'
    End
End

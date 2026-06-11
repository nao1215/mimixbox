Describe 'chpasswd'
    Include loginutils/chpasswd_test.sh

    It 'rejects an unknown method'
        When call TestChpasswdBadMethod
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestChpasswdHelp
        The status should be success
        The output should include 'Usage: chpasswd'
        The output should include 'password'
    End
End

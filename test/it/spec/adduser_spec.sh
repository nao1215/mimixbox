Describe 'adduser'
    Include loginutils/adduser_test.sh

    It 'requires a user name'
        When call TestAdduserNoName
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestAdduserHelp
        The status should be success
        The output should include 'Usage: adduser'
        The output should include 'account'
    End
End

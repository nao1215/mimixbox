Describe 'deluser'
    Include loginutils/deluser_test.sh

    It 'requires a user name'
        When call TestDeluserNoName
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestDeluserHelp
        The status should be success
        The output should include 'Usage: deluser'
        The output should include 'user'
    End
End

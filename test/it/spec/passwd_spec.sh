Describe 'passwd'
    Include loginutils/passwd_test.sh

    It 'rejects conflicting flags'
        When call TestPasswdConflict
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestPasswdHelp
        The status should be success
        The output should include 'Usage: passwd'
        The output should include 'password'
    End
End

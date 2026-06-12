Describe 'vlock'
    Include loginutils/vlock_test.sh

    It 'fails on a wrong password'
        When call TestVlockWrongPw
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestVlockHelp
        The status should be success
        The output should include 'Usage: vlock'
        The output should include 'terminal'
    End
End

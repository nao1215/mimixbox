Describe 'chsh'
    Include loginutils/chsh_test.sh

    It 'describes itself with --help'
        When call TestChshHelp
        The status should be success
        The output should include 'Usage: chsh'
        The output should include 'login shell'
    End
    It 'lists shells and exits successfully'
        When call TestChshListShells
        The output should equal 'rc=0'
        The status should be success
    End
    It 'rejects an unknown user'
        When call TestChshUnknownUser
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects a relative shell path'
        When call TestChshRelativeShell
        The output should equal 'rc=1'
        The status should be success
    End
End

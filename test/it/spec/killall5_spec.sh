Describe 'killall5'
    Include procps/killall5_test.sh

    It 'describes itself with --help'
        When call TestKillall5Help
        The status should be success
        The output should include 'Usage: killall5'
        The output should include 'signal'
    End
End

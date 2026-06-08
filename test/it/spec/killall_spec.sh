Describe 'killall'
    Include shellutils/killall_test.sh
    It 'kills a process by name'
        When call TestKillall
        The output should equal 'killed:0'
        The status should be success
    End
End

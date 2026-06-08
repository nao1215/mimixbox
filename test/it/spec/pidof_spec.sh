Describe 'pidof'
    Include shellutils/pidof_test.sh
    It 'finds the PID of a running process'
        When call TestPidofInit
        The output should equal 'found'
        The status should be success
    End
End

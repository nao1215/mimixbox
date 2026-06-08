Describe 'tty not a tty'
    Include shellutils/tty_test.sh
    It 'reports not a tty when stdin is a pipe'
        When call TestTtyNotATty
        The output should equal 'not a tty'
        The status should be failure
    End
End

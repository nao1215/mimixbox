Describe 'watch'
    Include shellutils/watch_test.sh
    It 'runs the command and shows its output'
        When call TestWatchOnce
        The output should include 'tick'
        The status should be failure
    End
End

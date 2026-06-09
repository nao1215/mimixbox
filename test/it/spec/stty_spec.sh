Describe 'stty'
    Include console-tools/stty_test.sh

    It 'reports when standard input is not a terminal'
        When call TestSttyNotTty
        The output should equal 'stty: standard input: not a tty
exit=1'
        The status should be success
    End
End

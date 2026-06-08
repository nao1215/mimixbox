Describe 'cowthink'
    Include jokeutils/cowthink_test.sh
    It 'draws the thought-bubble connector'
        When call TestCowthink
        The output should equal '   o'
        The status should be success
    End
End

Describe 'fortune'
    Include jokeutils/fortune_test.sh
    It 'prints a single adage line'
        When call TestFortune
        The output should equal '1'
        The status should be success
    End
End

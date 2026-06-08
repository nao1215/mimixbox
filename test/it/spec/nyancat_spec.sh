Describe 'nyancat'
    Include jokeutils/nyancat_test.sh
    It 'exits gracefully without a terminal'
        When call TestNyancatNoTTY
        The output should equal 'rc:0'
        The status should be success
    End
End

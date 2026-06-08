Describe 'cmatrix'
    Include jokeutils/cmatrix_test.sh
    It 'exits gracefully without a terminal'
        When call TestCmatrixNoTTY
        The output should equal 'rc:0'
        The status should be success
    End
End

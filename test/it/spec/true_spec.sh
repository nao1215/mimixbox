Describe 'True is True'
    Include shellutils/true_test.sh
    It 'print nothing, and exit-status is 0'
        When call TestTrue
        The output should equal ''
        The status should be success
    End
End
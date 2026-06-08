Describe 'nproc'
    Include shellutils/nproc_test.sh
    It 'prints a positive number'
        When call TestNproc
        The output should not equal ''
        The status should be success
    End
End

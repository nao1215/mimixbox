Describe 'rev from pipe'
    Include textutils/rev_test.sh
    It 'reverses the characters of a line'
        When call TestRevPipe
        The output should equal 'cba'
        The status should be success
    End
End

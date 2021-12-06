Describe 'False is False'
    Include shellutils/false_test.sh
    It 'print nothing, and exit-status is 1'
        When call TestFalse
        The output should equal ''
        The status should be failure
    End
End
Describe 'tr'
    Include textutils/tr_test.sh
    It 'translates lowercase to uppercase'
        When call TestTr
        The output should equal 'ABC'
        The status should be success
    End
End

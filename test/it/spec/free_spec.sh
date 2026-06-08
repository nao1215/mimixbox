Describe 'free'
    Include shellutils/free_test.sh
    It 'prints the column header'
        When call TestFreeHeader
        The output should include 'total'
        The output should include 'available'
        The status should be success
    End
End

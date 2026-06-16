Describe 'tr --truncate-set1'
    Include textutils/tr_truncate_test.sh

    It 'truncates SET1 to SET2 length, leaving extra chars unchanged'
        When call TestTrTruncateSet1
        The output should equal 'xyc'
        The status should be success
    End

    It 'accepts the -t short form'
        When call TestTrTruncateSet1Short
        The output should equal 'xyc'
        The status should be success
    End
End

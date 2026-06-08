Describe 'awk'
    Include editors/awk_test.sh

    It 'prints a field'
        When call TestAwkField
        The output should equal 'two'
        The status should be success
    End
    It 'honors -F'
        When call TestAwkFS
        The output should equal 'root'
        The status should be success
    End
    It 'selects a record with NR'
        When call TestAwkNR
        The output should equal 'b'
        The status should be success
    End
    It 'counts records in END'
        When call TestAwkEnd
        The output should equal '3'
        The status should be success
    End
End

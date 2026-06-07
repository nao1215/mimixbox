Describe 'hostid output format'
    Include shellutils/hostid_test.sh

    It 'prints 8 hexadecimal digits'
        When call TestHostidFormat
        The output should match pattern "[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]*"
        The status should be success
    End
End

Describe 'hostid is stable'
    Include shellutils/hostid_test.sh

    It 'prints the same value on repeated calls'
        When call TestHostidStable
        The output should equal 'stable'
        The status should be success
    End
End

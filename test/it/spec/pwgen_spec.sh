Describe 'pwgen'
    Include securityutils/pwgen_test.sh
    It 'generates the requested number of passwords'
        When call TestPwgenCount
        The output should equal '3'
        The status should be success
    End
End

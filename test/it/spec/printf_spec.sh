Describe 'printf'
    Include shellutils/printf_test.sh
    It 'formats arguments'
        When call TestPrintf
        The output should equal 'foo-bar'
        The status should be success
    End
End

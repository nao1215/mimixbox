Describe 'getty'
    Include loginutils/getty_test.sh

    It 'prints the login prompt'
        When call TestGettyPrompt
        The output should equal '1'
        The status should be success
    End
    It 'requires a TTY argument'
        When call TestGettyNoTTY
        The output should equal 'rc=1'
        The status should be success
    End
End

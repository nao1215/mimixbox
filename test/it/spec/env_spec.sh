Describe 'env sets a variable'
    Include shellutils/env_test.sh

    It 'adds the assignment to the printed environment'
        When call TestEnvAssign
        The output should equal 'FOO=bar'
        The status should be success
    End
End

Describe 'env -i starts from an empty environment'
    Include shellutils/env_test.sh

    It 'prints only the given assignment'
        When call TestEnvIgnore
        The output should equal 'ONLY=here'
        The status should be success
    End
End

Describe 'env runs a command with the modified environment'
    Include shellutils/env_test.sh

    It 'passes the variable to the command'
        When call TestEnvRunCommand
        The output should equal 'hi'
        The status should be success
    End
End

Describe 'Echo text without variable'
    Include shellutils/echo_test.sh
    It 'says Hello World!'
        When call TestEchoNormal
        The output should equal 'Hello World!'
    End
End

Describe 'Echo text with variable'
    Include shellutils/echo_test.sh
    It 'says Hello World! $1=World!'
        When call TestEchoNormal "World!"
        The output should equal 'Hello World!'
    End
End

Describe 'Echo text with environment variable'
    Include shellutils/echo_test.sh
    It 'says ${TEST_ENV}=TEST_ENV_VAR'
        When call TestEchoEnvVariable
        The output should equal 'TEST_ENV_VAR'
    End
End

Describe 'Echo pipe data without xargs command'
    Include shellutils/echo_test.sh
    It 'says nothing'
        When call TestEchoPipeWithoutXargs
        The output should equal ''
    End
End

Describe 'Echo pipe data with xargs command'
    Include shellutils/echo_test.sh
    It 'says pipe'
        When call TestEchoPipeWithargs
        The output should equal 'pipe'
    End
End

Describe 'Echo with no arguments'
    Include shellutils/echo_test.sh
    It 'says nothing'
        When call TestEchoNoArg
        The output should equal ''
    End
End

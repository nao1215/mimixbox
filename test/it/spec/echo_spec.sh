Describe 'Echo text without variable'
    Include shellutils/echo_test.sh
    It 'says Hello World!'
        When call TestEchoNormal
        The output should equal 'Hello World!'
        The status should be success
    End
End

Describe 'Echo text with variable'
    Include shellutils/echo_test.sh
    It 'says Hello World! $1=World!'
        When call TestEchoNormal "World!"
        The output should equal 'Hello World!'
        The status should be success
    End
End

Describe 'Echo text with environment variable'
    Include shellutils/echo_test.sh
    It 'says ${TEST_ENV}=TEST_ENV_VAR'
        When call TestEchoEnvVariable
        The output should equal 'TEST_ENV_VAR'
        The status should be success
    End
End

Describe 'Echo pipe data with xargs command'
    Include shellutils/echo_test.sh
    It 'says pipe'
        When call TestEchoPipeWithargs
        The output should equal 'pipe'
        The status should be success
    End
End

Describe 'Echo with no arguments'
    Include shellutils/echo_test.sh
    It 'says nothing'
        When call TestEchoNoArg
        The output should equal ''
        The status should be success
    End
End

Describe 'Echo redirect to file.'
    Include shellutils/echo_test.sh
    setup() { mkdir -p /tmp/it/mimixbox; }
    cleanup() { rm /tmp/it/mimixbox/echo.txt; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'redirect data to file and show it.'
        When call TestEchoRedirect
        The output should equal 'MimixBox'
        The status should be success
    End
End

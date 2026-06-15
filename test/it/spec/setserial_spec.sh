Describe 'setserial'
    Include console-tools/setserial_test.sh

    It 'echoes the parsed request with -g'
        When call TestSetserialGetEcho
        The output should equal '1'
        The status should be success
    End
    It 'rejects an unknown parameter'
        When call TestSetserialBadParam
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a device'
        When call TestSetserialNoDevice
        The output should equal 'rc=1'
        The status should be success
    End
End

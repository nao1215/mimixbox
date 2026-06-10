Describe 'ps'
    Include procps/ps_test.sh

    It 'prints the standard header'
        When call TestPsHeader
        The output should equal '    PID TTY          TIME CMD'
        The status should be success
    End
    It 'lists running processes'
        When call TestPsHasInit
        The status should be success
        The output should not equal '0'
    End
End

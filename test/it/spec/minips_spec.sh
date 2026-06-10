Describe 'minips'
    Include procps/minips_test.sh

    It 'prints the PID/USER/COMMAND header'
        When call TestMinipsHeader
        The output should equal '1'
        The status should be success
    End
    It 'lists processes'
        When call TestMinipsRows
        The status should be success
        The output should not equal '0'
    End
End

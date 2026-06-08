Describe 'timeout finishes in time'
    Include shellutils/timeout_test.sh
    It 'runs the command to completion'
        When call TestTimeoutFinishes
        The output should equal 'done'
        The status should be success
    End
End

Describe 'timeout expires'
    Include shellutils/timeout_test.sh
    It 'returns exit code 124 on timeout'
        When call TestTimeoutExpires
        The output should equal 'exit:124'
        The status should be success
    End
End

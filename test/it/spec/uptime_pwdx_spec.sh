Describe 'uptime / pwdx'
    Include procps/uptime_pwdx_test.sh

    It 'uptime shows the load averages'
        When call TestUptime
        The output should equal '1'
        The status should be success
    End
    It 'pwdx prints a process working directory'
        When call TestPwdx
        The output should equal '1'
        The status should be success
    End
End

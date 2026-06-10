Describe 'syslogd'
    Include procps/syslogd_test.sh

    It 'describes itself with --help'
        When call TestSyslogdHelp
        The status should be success
        The output should include 'Usage: syslogd'
        The output should include 'log'
    End
End

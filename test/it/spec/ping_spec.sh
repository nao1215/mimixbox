Describe 'ping usage'
    Include netutils/ping_test.sh
    It 'reports an error when no host is given'
        When call TestPingUsage
        The output should include 'rc:1'
        The status should be success
    End
End

Describe 'whris usage'
    Include netutils/whris_test.sh
    It 'reports an error when no domain is given'
        When call TestWhrisUsage
        The output should include 'rc:1'
        The status should be success
    End
End

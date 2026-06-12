Describe 'softlimit'
    Include runit/softlimit_test.sh

    It 'runs a program under the limits'
        When call TestSoftlimitRuns
        The output should equal 'rc=0'
        The status should be success
    End
    It 'requires a program'
        When call TestSoftlimitNoProg
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'addgroup / delgroup'
    Include loginutils/group_test.sh

    It 'addgroup requires a group name'
        When call TestAddgroupNoName
        The output should equal 'rc=1'
        The status should be success
    End
    It 'delgroup requires a group name'
        When call TestDelgroupNoName
        The output should equal 'rc=1'
        The status should be success
    End
End

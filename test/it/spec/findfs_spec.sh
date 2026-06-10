Describe 'findfs'
    Include util-linux/findfs_test.sh

    It 'fails for an unknown label'
        When call TestFindfsMissing
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects a malformed tag'
        When call TestFindfsBadSpec
        The output should equal 'rc=1'
        The status should be success
    End
End

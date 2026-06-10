Describe 'switch_root'
    Include util-linux/switch_root_test.sh

    It 'requires NEW_ROOT and INIT'
        When call TestSwitchRootMissingInit
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects a non-directory NEW_ROOT'
        When call TestSwitchRootBadDir
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'fsfreeze'
    Include util-linux/fsfreeze_test.sh

    It 'requires a freeze or unfreeze mode'
        When call TestFsfreezeNoMode
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects both modes at once'
        When call TestFsfreezeBothModes
        The output should equal 'rc=1'
        The status should be success
    End
End

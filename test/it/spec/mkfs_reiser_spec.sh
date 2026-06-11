Describe 'mkfs.reiser'
    Include util-linux/mkfs_reiser_test.sh

    It 'refuses deterministically'
        When call TestMkfsReiserRefuses
        The output should equal 'rc=1'
        The status should be success
    End
    It 'explains that ReiserFS is deprecated'
        When call TestMkfsReiserMessage
        The output should equal '1'
        The status should be success
    End
End

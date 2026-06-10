Describe 'chattr'
    Include util-linux/chattr_test.sh

    It 'rejects a malformed mode'
        When call TestChattrBadMode
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects an unknown attribute'
        When call TestChattrBadAttr
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'chat'
    Include console-tools/chat_test.sh

    It 'sends the reply after the expected string'
        When call TestChatSendsReply
        The output should equal '1'
        The status should be success
    End
    It 'requires a script'
        When call TestChatNoScript
        The output should equal 'rc=1'
        The status should be success
    End
    It 'fails when an expected string never arrives'
        When call TestChatExpectNeverSeen
        The output should equal 'rc=1'
        The status should be success
    End
End

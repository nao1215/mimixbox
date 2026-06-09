Describe 'compat front-ends'
    Include compat/compat_test.sh

    It 'the [ alias returns true for an existing file'
        When call TestBracketTrue
        The output should equal 'yes'
        The status should be success
    End
    It 'the [ alias returns false for a missing file'
        When call TestBracketFalse
        The output should equal 'no'
        The status should be success
    End
    It 'busybox dispatches to an applet'
        When call TestBusyboxDispatch
        The output should equal 'hello'
        The status should be success
    End
    It 'busybox --list shows applets'
        When call TestBusyboxList
        The status should be success
        The output should include 'cat'
        The output should include 'busybox'
    End
    It 'sh -c runs a command without a prompt'
        When call TestShDashC
        The output should equal 'from-sh'
        The status should be success
    End
    It 'bash reads a non-interactive script from stdin without a prompt'
        When call TestBashStdinNoPrompt
        The output should equal 'ok'
        The status should be success
    End
End

Describe 'showkey'
    Include console-tools/showkey_test.sh

    It 'rejects conflicting modes'
        When call TestShowkeyConflict
        The output should equal 'rc=1'
        The status should be success
    End
    It 'fails deterministically without a console'
        When call TestShowkeyNoConsole
        The output should equal 'rc=1'
        The status should be success
    End
End

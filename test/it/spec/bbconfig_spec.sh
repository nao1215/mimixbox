Describe 'bbconfig'
    Include console-tools/bbconfig_test.sh

    It 'prints the version line'
        When call TestBbconfigHasVersionLine
        The output should equal '1'
        The status should be success
    End
    It 'lists itself among the applets'
        When call TestBbconfigListsItself
        The output should equal '1'
        The status should be success
    End
    It 'rejects an unexpected argument'
        When call TestBbconfigRejectsArg
        The output should equal 'rc=1'
        The status should be success
    End
End

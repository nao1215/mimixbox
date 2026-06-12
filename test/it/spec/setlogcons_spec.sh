Describe 'setlogcons'
    Include console-tools/setlogcons_test.sh

    It 'rejects a non-numeric VT'
        When call TestSetlogconsBadN
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestSetlogconsHelp
        The status should be success
        The output should include 'Usage: setlogcons'
        The output should include 'kernel'
    End
End

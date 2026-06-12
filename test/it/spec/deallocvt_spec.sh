Describe 'deallocvt'
    Include console-tools/deallocvt_test.sh

    It 'rejects a non-numeric VT'
        When call TestDeallocvtBadN
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestDeallocvtHelp
        The status should be success
        The output should include 'Usage: deallocvt'
        The output should include 'virtual terminal'
    End
End

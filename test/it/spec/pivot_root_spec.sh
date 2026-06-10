Describe 'pivot_root'
    Include util-linux/pivot_root_test.sh

    It 'requires two directories'
        When call TestPivotRootBadArgs
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestPivotRootHelp
        The status should be success
        The output should include 'Usage: pivot_root'
        The output should include 'root'
    End
End

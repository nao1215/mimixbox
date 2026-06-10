Describe 'lsattr'
    Include util-linux/lsattr_test.sh

    It 'describes itself with --help'
        When call TestLsattrHelp
        The status should be success
        The output should include 'Usage: lsattr'
        The output should include 'attribute'
    End
End

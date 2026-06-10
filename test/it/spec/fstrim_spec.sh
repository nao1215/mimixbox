Describe 'fstrim'
    Include util-linux/fstrim_test.sh

    It 'requires a mount point'
        When call TestFstrimNoArg
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFstrimHelp
        The status should be success
        The output should include 'Usage: fstrim'
        The output should include 'Discard'
    End
End

Describe 'blkdiscard'
    Include util-linux/blkdiscard_test.sh

    It 'requires a device'
        When call TestBlkdiscardNoDev
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestBlkdiscardHelp
        The status should be success
        The output should include 'Usage: blkdiscard'
        The output should include 'Discard'
    End
End

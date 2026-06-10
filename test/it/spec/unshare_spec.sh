Describe 'unshare'
    Include util-linux/unshare_test.sh

    It 'requires a namespace flag'
        When call TestUnshareNoFlag
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestUnshareHelp
        The status should be success
        The output should include 'Usage: unshare'
        The output should include 'namespace'
    End
End

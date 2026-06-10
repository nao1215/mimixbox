Describe 'fbset'
    Include util-linux/fbset_test.sh

    It 'fails on a missing framebuffer'
        When call TestFbsetNoFb
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFbsetHelp
        The status should be success
        The output should include 'Usage: fbset'
        The output should include 'framebuffer'
    End
End

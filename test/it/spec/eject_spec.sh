Describe 'eject'
    Include util-linux/eject_test.sh

    It 'describes itself with --help'
        When call TestEjectHelp
        The status should be success
        The output should include 'Usage: eject'
        The output should include 'media'
    End
    It 'fails on a missing device'
        When call TestEjectMissing
        The output should equal 'rc=1'
        The status should be success
    End
End

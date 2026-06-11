Describe 'uevent'
    Include util-linux/uevent_test.sh

    It 'describes itself with --help'
        When call TestUeventHelp
        The status should be success
        The output should include 'Usage: uevent'
        The output should include 'uevent'
    End
End

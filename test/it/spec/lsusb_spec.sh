Describe 'lsusb'
    Include util-linux/lsusb_test.sh

    It 'describes itself with --help'
        When call TestLsusbHelp
        The status should be success
        The output should include 'Usage: lsusb'
        The output should include 'USB'
    End
End

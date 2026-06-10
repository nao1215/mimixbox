Describe 'hwclock'
    Include util-linux/hwclock_test.sh

    It 'describes itself with --help'
        When call TestHwclockHelp
        The status should be success
        The output should include 'Usage: hwclock'
        The output should include 'RTC'
    End
End

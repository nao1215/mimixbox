Describe 'dmesg'
    Include util-linux/dmesg_test.sh

    It 'describes itself with --help'
        When call TestDmesgHelp
        The status should be success
        The output should include 'Usage: dmesg'
        The output should include 'kernel ring buffer'
    End
End

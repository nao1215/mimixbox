Describe 'fdflush'
    Include util-linux/fdflush_test.sh

    It 'requires a device'
        When call TestFdflushNoDev
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFdflushHelp
        The status should be success
        The output should include 'Usage: fdflush'
        The output should include 'floppy'
    End
End

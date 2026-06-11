Describe 'fdformat'
    Include util-linux/fdformat_test.sh

    It 'requires a device'
        When call TestFdformatNoDev
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFdformatHelp
        The status should be success
        The output should include 'Usage: fdformat'
        The output should include 'floppy'
    End
End

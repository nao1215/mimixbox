Describe 'fgconsole'
    Include console-tools/fgconsole_test.sh

    It 'fails without a virtual console'
        When call TestFgconsoleNoVT
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFgconsoleHelp
        The status should be success
        The output should include 'Usage: fgconsole'
        The output should include 'virtual terminal'
    End
End

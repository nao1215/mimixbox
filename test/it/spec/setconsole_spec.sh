Describe 'setconsole'
    Include console-tools/setconsole_test.sh

    It 'fails on an inaccessible device'
        When call TestSetconsoleBadDev
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestSetconsoleHelp
        The status should be success
        The output should include 'Usage: setconsole'
        The output should include 'console'
    End
End

Describe 'runsvdir'
    Include runit/runsvdir_test.sh

    It 'requires a services directory'
        When call TestRunsvdirNoDir
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestRunsvdirHelp
        The status should be success
        The output should include 'Usage: runsvdir'
        The output should include 'service'
    End
End

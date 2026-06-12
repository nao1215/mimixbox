Describe 'runsv'
    Include runit/runsv_test.sh

    It 'requires a service directory'
        When call TestRunsvNoDir
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestRunsvHelp
        The status should be success
        The output should include 'Usage: runsv'
        The output should include 'service'
    End
End

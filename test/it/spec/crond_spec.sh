Describe 'crond'
    Include loginutils/crond_test.sh

    It 'requires foreground mode'
        When call TestCrondNoForeground
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestCrondHelp
        The status should be success
        The output should include 'Usage: crond'
        The output should include 'cron'
    End
End

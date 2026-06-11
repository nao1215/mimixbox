Describe 'crontab'
    Include loginutils/crontab_test.sh

    It 'reports that interactive edit is unsupported'
        When call TestCrontabEdit
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestCrontabHelp
        The status should be success
        The output should include 'Usage: crontab'
        The output should include 'crontab'
    End
End

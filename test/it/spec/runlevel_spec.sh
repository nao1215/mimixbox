Describe 'runlevel'
    Include loginutils/runlevel_test.sh

    It 'reports a runlevel or unknown'
        When call TestRunlevelRuns
        The line 1 of output should match pattern '*'
        The line 2 of output should match pattern 'rc=*'
    End
End

Describe 'pgrep / pkill'
    Include procps/pgrep_test.sh

    It 'finds a running process by name'
        When call TestPgrepFindsSleep
        The output should equal '1'
        The status should be success
    End
    It 'exits non-zero when nothing matches'
        When call TestPgrepNoMatch
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'kill'
    Include shellutils/kill_test.sh
    It 'lists signal names with -l'
        When call TestKillList
        The output should equal 'ok'
        The status should be success
    End
End

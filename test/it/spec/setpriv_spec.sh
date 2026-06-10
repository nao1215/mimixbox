Describe 'setpriv'
    Include util-linux/setpriv_test.sh

    It 'dumps the current privileges'
        When call TestSetprivDump
        The status should be success
        The output should not equal '0'
    End
    It 'runs a command with --no-new-privs'
        When call TestSetprivRun
        The output should equal 'ran'
        The status should be success
    End
End

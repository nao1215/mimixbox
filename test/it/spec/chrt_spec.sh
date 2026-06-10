Describe 'chrt'
    Include util-linux/chrt_test.sh

    It 'prints a process scheduling policy'
        When call TestChrtPrint
        The output should equal '1'
        The status should be success
    End
    It 'runs a command under a scheduling policy'
        When call TestChrtRun
        The output should equal 'scheduled'
        The status should be success
    End
End

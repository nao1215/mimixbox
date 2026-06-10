Describe 'ionice'
    Include util-linux/ionice_test.sh

    It 'prints a process I/O class'
        When call TestIonicePrint
        The output should equal '1'
        The status should be success
    End
    It 'runs a command at a given I/O class'
        When call TestIoniceRun
        The output should equal 'idled'
        The status should be success
    End
End

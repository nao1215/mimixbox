Describe 'taskset'
    Include util-linux/taskset_test.sh

    It 'prints a process affinity mask'
        When call TestTasksetPrint
        The output should equal '1'
        The status should be success
    End
    It 'runs a command bound to a CPU'
        When call TestTasksetRun
        The output should equal 'affined'
        The status should be success
    End
    It 'rejects an invalid mask'
        When call TestTasksetInvalid
        The output should equal 'rc=1'
        The status should be success
    End
End

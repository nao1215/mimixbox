Describe 'swapon / swapoff'
    Include util-linux/swap_test.sh

    It 'swapon -s prints the swaps header'
        When call TestSwaponSummary
        The output should equal '1'
        The status should be success
    End
    It 'swapoff requires a target'
        When call TestSwapoffNoArg
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'lspci'
    Include util-linux/lspci_test.sh

    It 'lists PCI devices and exits successfully'
        When call TestLspciRuns
        The output should equal '0'
        The status should be success
    End
End

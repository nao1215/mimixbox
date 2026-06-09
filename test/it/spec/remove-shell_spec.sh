Describe 'remove-shell CLI contract'
    Include debianutils/remove-shell_test.sh
    It 'prints usage with --help and exits 0'
        When call RemoveShellHelp
        The status should be success
        The output should include 'Usage: remove-shell'
    End
    It 'fails with a message when given no operand'
        When call RemoveShellNoArg
        The status should be failure
        The error should include 'remove-shell'
    End
End

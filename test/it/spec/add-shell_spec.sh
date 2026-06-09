Describe 'add-shell CLI contract'
    Include debianutils/add-shell_test.sh
    It 'prints usage with --help and exits 0'
        When call AddShellHelp
        The status should be success
        The output should include 'Usage: add-shell'
    End
    It 'fails with a message when given no operand'
        When call AddShellNoArg
        The status should be failure
        The error should include 'add-shell'
    End
End

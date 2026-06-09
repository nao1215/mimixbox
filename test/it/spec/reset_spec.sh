Describe 'reset CLI contract'
    Include console-tools/reset_test.sh
    It 'prints usage with --help and exits 0'
        When call ResetHelp
        The status should be success
        The output should include 'Usage: reset'
    End
End

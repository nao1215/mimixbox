Describe 'clear CLI contract'
    Include console-tools/clear_test.sh
    It 'prints usage with --help and exits 0'
        When call ClearHelp
        The status should be success
        The output should include 'Usage: clear'
    End
    It 'exits 0 when clearing the screen'
        When call ClearRun
        The status should be success
    End
End

Describe 'wget CLI contract'
    Include shellutils/wget_test.sh
    It 'prints usage with --help and exits 0'
        When call WgetHelp
        The status should be success
        The output should include 'Usage: wget'
    End
    It 'fails with a message when given no operand'
        When call WgetNoArg
        The status should be failure
        The error should include 'wget'
    End
End

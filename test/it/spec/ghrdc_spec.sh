Describe 'ghrdc CLI contract'
    Include shellutils/ghrdc_test.sh
    It 'prints usage with --help and exits 0'
        When call GhrdcHelp
        The status should be success
        The output should include 'Usage: ghrdc'
    End
    It 'fails with a message when given no operand'
        When call GhrdcNoArg
        The status should be failure
        The error should include 'ghrdc'
    End
End

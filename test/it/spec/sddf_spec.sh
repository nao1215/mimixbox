Describe 'sddf CLI contract'
    Include shellutils/sddf_contract_test.sh
    It 'prints usage with --help and exits 0'
        When call SddfHelp
        The status should be success
        The output should include 'Usage: sddf'
    End
    It 'fails with a message when given no operand'
        When call SddfNoArg
        The status should be failure
        The error should include 'sddf'
    End
End

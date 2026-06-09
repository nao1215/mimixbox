Describe 'chown CLI contract'
    Include fileutils/chown_contract_test.sh
    It 'prints usage with --help and exits 0'
        When call ChownHelp
        The status should be success
        The output should include 'Usage: chown'
    End
    It 'fails with a message when given no operand'
        When call ChownNoArg
        The status should be failure
        The error should include 'chown'
    End
End

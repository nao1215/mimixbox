Describe 'chgrp CLI contract'
    Include fileutils/chgrp_contract_test.sh
    It 'prints usage with --help and exits 0'
        When call ChgrpHelp
        The status should be success
        The output should include 'Usage: chgrp'
    End
    It 'fails with a message when given no operand'
        When call ChgrpNoArg
        The status should be failure
        The error should include 'chgrp'
    End
End

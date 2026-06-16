# Per-command --help contract specs backed by dedicated textutils helpers (issue #489).
Describe 'textutils commands expose a dedicated --help helper'
    Include textutils/crc32_test.sh
    Include textutils/sha384sum_test.sh
    Include textutils/sha3sum_test.sh
    Include textutils/sum_test.sh
    Include textutils/uudecode_test.sh
    Include textutils/uuencode_test.sh

    It 'crc32 --help is structured'
        When call Crc32Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'sha384sum --help is structured'
        When call Sha384sumHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sha3sum --help is structured'
        When call Sha3sumHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sum --help is structured'
        When call SumHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'uudecode --help is structured'
        When call UudecodeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'uuencode --help is structured'
        When call UuencodeHelp
        The status should be success
        The output should include 'Usage:'
    End
End

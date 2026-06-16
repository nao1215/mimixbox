# Per-command --help contract specs backed by dedicated printutils helpers (issue #489).
Describe 'printutils commands expose a dedicated --help helper'
    Include printutils/lpd_test.sh
    Include printutils/lpq_test.sh
    Include printutils/lpr_test.sh

    It 'lpd --help is structured'
        When call LpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lpq --help is structured'
        When call LpqHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lpr --help is structured'
        When call LprHelp
        The status should be success
        The output should include 'Usage:'
    End
End

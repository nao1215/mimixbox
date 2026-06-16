# Per-command --help contract specs backed by dedicated findutils helpers (issue #489).
Describe 'findutils commands expose a dedicated --help helper'
    Include findutils/egrep_test.sh
    Include findutils/fgrep_test.sh

    It 'egrep --help is structured'
        When call EgrepHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'fgrep --help is structured'
        When call FgrepHelp
        The status should be success
        The output should include 'Usage:'
    End
End

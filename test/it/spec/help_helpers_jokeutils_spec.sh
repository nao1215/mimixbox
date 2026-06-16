# Per-command --help contract specs backed by dedicated jokeutils helpers (issue #489).
Describe 'jokeutils commands expose a dedicated --help helper'
    Include jokeutils/sl_test.sh

    It 'sl --help is structured'
        When call SlHelp
        The status should be success
        The output should include 'Usage:'
    End
End

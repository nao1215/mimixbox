# Per-command --help contract specs backed by dedicated pmutils helpers (issue #489).
Describe 'pmutils commands expose a dedicated --help helper'
    Include pmutils/poweroff_test.sh
    Include pmutils/reboot_test.sh

    It 'poweroff --help is structured'
        When call PoweroffHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'reboot --help is structured'
        When call RebootHelp
        The status should be success
        The output should include 'Usage:'
    End
End

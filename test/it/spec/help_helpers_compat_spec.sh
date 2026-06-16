# Per-command --help contract specs backed by dedicated compat helpers (issue #489).
Describe 'compat commands expose a dedicated --help helper'
    Include compat/[_test.sh
    Include compat/[[_test.sh
    Include compat/ash_test.sh
    Include compat/bash_test.sh
    Include compat/busybox_test.sh
    Include compat/cttyhack_test.sh
    Include compat/hush_test.sh
    Include compat/unit_test.sh

    It '[ --help is structured'
        When call BracketHelp
        The status should be success
        The output should include 'Usage:'
    End
    It '[[ --help is structured'
        When call DoubleBracketHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ash --help is structured'
        When call AshHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'bash --help is structured'
        When call BashHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'busybox --help is structured'
        When call BusyboxHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'cttyhack --help is structured'
        When call CttyhackHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'hush --help is structured'
        When call HushHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'unit --help is structured'
        When call UnitHelp
        The status should be success
        The output should include 'Usage:'
    End
End

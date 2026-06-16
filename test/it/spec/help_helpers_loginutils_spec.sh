# Per-command --help contract specs backed by dedicated loginutils helpers (issue #489).
Describe 'loginutils commands expose a dedicated --help helper'
    Include loginutils/addgroup_test.sh
    Include loginutils/delgroup_test.sh
    Include loginutils/linuxrc_test.sh
    Include loginutils/run-init_test.sh
    Include loginutils/run-parts_test.sh
    Include loginutils/start-stop-daemon_test.sh

    It 'addgroup --help is structured'
        When call AddgroupHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'delgroup --help is structured'
        When call DelgroupHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'linuxrc --help is structured'
        When call LinuxrcHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'run-init --help is structured'
        When call RunInitHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'run-parts --help is structured'
        When call RunPartsHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'start-stop-daemon --help is structured'
        When call StartStopDaemonHelp
        The status should be success
        The output should include 'Usage:'
    End
End

# Per-command --help contract specs backed by dedicated procps helpers (issue #489).
Describe 'procps commands expose a dedicated --help helper'
    Include procps/depmod_test.sh
    Include procps/insmod_test.sh
    Include procps/lsmod_test.sh
    Include procps/modinfo_test.sh
    Include procps/modprobe_test.sh
    Include procps/pkill_test.sh
    Include procps/pwdx_test.sh
    Include procps/rmmod_test.sh
    Include procps/uptime_test.sh

    It 'depmod --help is structured'
        When call DepmodHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'insmod --help is structured'
        When call InsmodHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lsmod --help is structured'
        When call LsmodHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'modinfo --help is structured'
        When call ModinfoHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'modprobe --help is structured'
        When call ModprobeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'pkill --help is structured'
        When call PkillHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'pwdx --help is structured'
        When call PwdxHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'rmmod --help is structured'
        When call RmmodHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'uptime --help is structured'
        When call UptimeHelp
        The status should be success
        The output should include 'Usage:'
    End
End

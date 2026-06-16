# Per-command --help contract specs backed by dedicated console-tools helpers (issue #489).
Describe 'console-tools commands expose a dedicated --help helper'
    Include console-tools/adjtimex_test.sh
    Include console-tools/conspy_test.sh
    Include console-tools/dumpkmap_test.sh
    Include console-tools/less_test.sh
    Include console-tools/loadfont_test.sh
    Include console-tools/loadkmap_test.sh
    Include console-tools/microcom_test.sh
    Include console-tools/more_test.sh
    Include console-tools/openvt_test.sh
    Include console-tools/rx_test.sh
    Include console-tools/setfont_test.sh

    It 'adjtimex --help is structured'
        When call AdjtimexHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'conspy --help is structured'
        When call ConspyHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dumpkmap --help is structured'
        When call DumpkmapHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'less --help is structured'
        When call LessHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'loadfont --help is structured'
        When call LoadfontHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'loadkmap --help is structured'
        When call LoadkmapHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'microcom --help is structured'
        When call MicrocomHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'more --help is structured'
        When call MoreHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'openvt --help is structured'
        When call OpenvtHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'rx --help is structured'
        When call RxHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'setfont --help is structured'
        When call SetfontHelp
        The status should be success
        The output should include 'Usage:'
    End
End

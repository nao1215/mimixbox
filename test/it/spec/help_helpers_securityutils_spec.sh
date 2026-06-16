# Per-command --help contract specs backed by dedicated securityutils helpers (issue #489).
Describe 'securityutils commands expose a dedicated --help helper'
    Include securityutils/chcon_test.sh
    Include securityutils/getenforce_test.sh
    Include securityutils/getsebool_test.sh
    Include securityutils/load_policy_test.sh
    Include securityutils/matchpathcon_test.sh
    Include securityutils/restorecon_test.sh
    Include securityutils/runcon_test.sh
    Include securityutils/selinuxenabled_test.sh
    Include securityutils/sestatus_test.sh
    Include securityutils/setenforce_test.sh
    Include securityutils/setfiles_test.sh
    Include securityutils/setsebool_test.sh
    Include securityutils/zip-pwcrack_test.sh

    It 'chcon --help is structured'
        When call ChconHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'getenforce --help is structured'
        When call GetenforceHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'getsebool --help is structured'
        When call GetseboolHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'load_policy --help is structured'
        When call LoadPolicyHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'matchpathcon --help is structured'
        When call MatchpathconHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'restorecon --help is structured'
        When call RestoreconHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'runcon --help is structured'
        When call RunconHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'selinuxenabled --help is structured'
        When call SelinuxenabledHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sestatus --help is structured'
        When call SestatusHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'setenforce --help is structured'
        When call SetenforceHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'setfiles --help is structured'
        When call SetfilesHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'setsebool --help is structured'
        When call SetseboolHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'zip-pwcrack --help is structured'
        When call ZipPwcrackHelp
        The status should be success
        The output should include 'Usage:'
    End
End

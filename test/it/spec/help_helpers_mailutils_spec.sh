# Per-command --help contract specs backed by dedicated mailutils helpers (issue #489).
Describe 'mailutils commands expose a dedicated --help helper'
    Include mailutils/makemime_test.sh
    Include mailutils/popmaildir_test.sh
    Include mailutils/reformime_test.sh
    Include mailutils/sendmail_test.sh

    It 'makemime --help is structured'
        When call MakemimeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'popmaildir --help is structured'
        When call PopmaildirHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'reformime --help is structured'
        When call ReformimeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sendmail --help is structured'
        When call SendmailHelp
        The status should be success
        The output should include 'Usage:'
    End
End

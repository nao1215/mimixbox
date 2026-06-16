# Per-command --help contract specs backed by dedicated shellutils helpers (issue #489).
Describe 'shellutils commands expose a dedicated --help helper'
    Include shellutils/fsync_test.sh
    Include shellutils/log-collect_test.sh
    Include shellutils/sddf_test.sh
    Include shellutils/time_test.sh
    Include shellutils/usleep_test.sh

    It 'fsync --help is structured'
        When call FsyncHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'log-collect --help is structured'
        When call LogCollectHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sddf --help is structured'
        When call SddfHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'time --help is structured'
        When call TimeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'usleep --help is structured'
        When call UsleepHelp
        The status should be success
        The output should include 'Usage:'
    End
End

# Per-command --help contract specs backed by dedicated embedded helpers (issue #489).
Describe 'embedded commands expose a dedicated --help helper'
    Include embedded/devmem_test.sh
    Include embedded/i2cdetect_test.sh
    Include embedded/i2cdump_test.sh
    Include embedded/i2cget_test.sh
    Include embedded/i2cset_test.sh
    Include embedded/partprobe_test.sh
    Include embedded/raidautorun_test.sh
    Include embedded/readahead_test.sh
    Include embedded/resume_test.sh
    Include embedded/seedrng_test.sh
    Include embedded/volname_test.sh
    Include embedded/watchdog_test.sh

    It 'devmem --help is structured'
        When call DevmemHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'i2cdetect --help is structured'
        When call I2cdetectHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'i2cdump --help is structured'
        When call I2cdumpHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'i2cget --help is structured'
        When call I2cgetHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'i2cset --help is structured'
        When call I2csetHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'partprobe --help is structured'
        When call PartprobeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'raidautorun --help is structured'
        When call RaidautorunHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'readahead --help is structured'
        When call ReadaheadHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'resume --help is structured'
        When call ResumeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'seedrng --help is structured'
        When call SeedrngHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'volname --help is structured'
        When call VolnameHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'watchdog --help is structured'
        When call WatchdogHelp
        The status should be success
        The output should include 'Usage:'
    End
End

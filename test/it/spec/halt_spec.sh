Describe 'halt --help'
    Include pmutils/halt_test.sh

    It 'prints usage and lists the options'
        When call TestHaltHelp
        The output should include 'Usage: halt'
        The output should include '--poweroff'
        The output should include '--wtmp-only'
        The status should be success
    End
End

Describe 'poweroff --help'
    Include pmutils/halt_test.sh

    It 'prints usage'
        When call TestPoweroffHelp
        The output should include 'Usage: poweroff'
        The status should be success
    End
End

Describe 'reboot --help'
    Include pmutils/halt_test.sh

    It 'prints usage'
        When call TestRebootHelp
        The output should include 'Usage: reboot'
        The status should be success
    End
End

Describe 'halt --version'
    Include pmutils/halt_test.sh

    It 'prints the version'
        When call TestHaltVersion
        The output should include 'halt (mimixbox)'
        The status should be success
    End
End

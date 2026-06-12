Describe 'acpid'
    Include loginutils/acpid_test.sh

    It 'requires foreground mode'
        When call TestAcpidNoForeground
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestAcpidHelp
        The status should be success
        The output should include 'Usage: acpid'
        The output should include 'ACPI'
    End
End

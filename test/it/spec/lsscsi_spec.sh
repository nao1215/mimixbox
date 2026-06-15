Describe 'lsscsi'
    Include embedded/lsscsi_test.sh

    It 'lists SCSI devices from sysfs without error (empty is allowed)'
        When call TestLsscsiRuns
        The output should equal 'ok'
        The status should be success
    End

    It 'prints usage for --help'
        When call TestLsscsiHelp
        The output should include 'Usage: lsscsi'
        The status should be success
    End

    It 'prints the version line for --version'
        When call TestLsscsiVersion
        The output should include 'lsscsi (mimixbox)'
        The status should be success
    End
End

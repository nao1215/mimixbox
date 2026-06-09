Describe 'ischroot CLI contract'
    Include debianutils/ischroot_test.sh
    It 'prints usage with --help and exits 0'
        When call IschrootHelp
        The status should be success
        The output should include 'Usage: ischroot'
    End
End

Describe 'valid-shell CLI contract'
    Include debianutils/valid-shell_test.sh
    It 'prints usage with --help and exits 0'
        When call ValidShellHelp
        The status should be success
        The output should include 'Usage: valid-shell'
    End
    It 'accepts a file listing existing shells'
        When call ValidShellValidFile
        The status should be success
        The output should include 'OK'
    End
End

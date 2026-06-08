Describe 'hostname'
    Include shellutils/hostname_test.sh
    It 'prints a non-empty host name'
        When call TestHostname
        The output should not equal ''
        The status should be success
    End
End

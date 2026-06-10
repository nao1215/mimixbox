Describe 'klogd'
    Include procps/klogd_test.sh

    It 'describes itself with --help'
        When call TestKlogdHelp
        The status should be success
        The output should include 'Usage: klogd'
        The output should include 'kernel'
    End
End

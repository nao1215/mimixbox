Describe 'sysctl'
    Include procps/sysctl_test.sh

    It 'reads a kernel parameter'
        When call TestSysctlRead
        The output should equal 'kernel.ostype = Linux'
        The status should be success
    End
    It 'lists parameters with -a'
        When call TestSysctlAll
        The status should be success
        The output should not equal '0'
    End
End

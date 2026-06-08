Describe 'arch'
    Include shellutils/arch_test.sh
    It 'prints a non-empty machine name'
        When call TestArch
        The output should not equal ''
        The status should be success
    End
End

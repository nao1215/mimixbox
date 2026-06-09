Describe 'chroot CLI contract'
    Include shellutils/chroot_test.sh
    It 'prints usage with --help and exits 0'
        When call ChrootHelp
        The status should be success
        The output should include 'Usage: chroot'
    End
    It 'fails with a message when given no operand'
        When call ChrootNoArg
        The status should be failure
        The error should include 'chroot'
    End
End

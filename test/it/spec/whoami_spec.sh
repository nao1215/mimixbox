Describe 'whoami without options'
    Include shellutils/whoami_test.sh

    It 'prints the current user name'
        When call TestWhoamiPrintsUser
        The output should equal "$(id -un)"
        The status should be success
    End
End

Describe 'whoami with an extra operand'
    Include shellutils/whoami_test.sh

    It 'reports an error'
        When call TestWhoamiExtraOperand
        The error should equal "whoami: extra operand 'extra'"
        The status should be failure
    End
End

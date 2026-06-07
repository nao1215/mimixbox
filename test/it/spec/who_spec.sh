Describe 'who on an empty utmp'
    Include shellutils/who_test.sh

    It 'prints nothing and succeeds'
        When call TestWhoEmpty
        The output should equal 'rc=0'
        The status should be success
    End
End

Describe 'who -q on an empty utmp'
    Include shellutils/who_test.sh

    It 'reports zero users'
        When call TestWhoCount
        The output should equal '# users=0'
        The status should be success
    End
End

Describe 'who --help'
    Include shellutils/who_test.sh

    It 'prints usage'
        When call TestWhoHelp
        The output should include 'Usage: who'
        The status should be success
    End
End

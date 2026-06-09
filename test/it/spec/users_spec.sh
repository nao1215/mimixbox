Describe 'users'
    Include shellutils/users_test.sh

    It 'runs and exits successfully'
        When call TestUsersExit
        The output should equal '0'
        The status should be success
    End
    It 'treats a missing utmp as nobody logged in'
        When call TestUsersMissing
        The output should equal '[] rc=0'
        The status should be success
    End
End

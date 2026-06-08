Describe 'pwd'
    Include shellutils/pwd_test.sh
    It 'prints the working directory'
        When call TestPwd
        The output should equal '/tmp'
        The status should be success
    End
End

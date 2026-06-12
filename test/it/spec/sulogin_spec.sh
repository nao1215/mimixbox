Describe 'sulogin'
    Include loginutils/sulogin_test.sh

    It 'rejects a wrong root password'
        When call TestSuloginWrongPw
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestSuloginHelp
        The status should be success
        The output should include 'Usage: sulogin'
        The output should include 'root'
    End
End

Describe 'fuser'
    Include procps/fuser_test.sh

    It 'finds processes using the current directory'
        When call TestFuserCwd
        The status should be success
        The output should equal '1'
    End
    It 'exits non-zero when nothing uses the file'
        When call TestFuserNone
        The output should equal 'rc=1'
        The status should be success
    End
End

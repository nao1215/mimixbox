Describe 'setuidgid'
    Include runit/setuidgid_test.sh

    It 'fails for an unknown user'
        When call TestSetuidgidUnknown
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a program'
        When call TestSetuidgidNoProg
        The output should equal 'rc=1'
        The status should be success
    End
End

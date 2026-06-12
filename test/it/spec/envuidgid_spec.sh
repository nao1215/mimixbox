Describe 'envuidgid'
    Include runit/envuidgid_test.sh

    It 'exports root uid and gid'
        When call TestEnvuidgidRoot
        The output should equal '0:0'
        The status should be success
    End
    It 'fails for an unknown user'
        When call TestEnvuidgidUnknown
        The output should equal 'rc=1'
        The status should be success
    End
End

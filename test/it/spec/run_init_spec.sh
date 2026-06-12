Describe 'run-init'
    Include loginutils/run_init_test.sh

    It 'requires NEW_ROOT and INIT'
        When call TestRunInitMissingInit
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects a non-directory NEW_ROOT'
        When call TestRunInitBadDir
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'nologin'
    Include loginutils/nologin_test.sh

    It 'prints a refusal and exits non-zero'
        When call TestNologinRefuses
        The line 1 of output should include 'not available'
        The line 2 of output should equal 'rc=1'
    End
    It 'never runs a passed command'
        When call TestNologinIgnoresArgs
        The output should equal '0'
        The status should be success
    End
End

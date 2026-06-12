Describe 'beep'
    Include console-tools/beep_test.sh

    It 'rejects a non-positive frequency'
        When call TestBeepBadFreq
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects a zero repeat count'
        When call TestBeepBadRepeat
        The output should equal 'rc=1'
        The status should be success
    End
End

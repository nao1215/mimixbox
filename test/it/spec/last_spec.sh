Describe 'last'
    Include util-linux/last_test.sh

    It 'treats an empty wtmp as no history and exits 0'
        When call TestLastRuns
        The output should equal '[] rc=0'
        The status should be success
    End
    It 'fails on a missing wtmp file'
        When call TestLastMissing
        The output should equal 'rc=1'
        The status should be success
    End
End

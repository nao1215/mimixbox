Describe 'shuf range'
    Include textutils/shuf_test.sh
    It 'shuffles a single-element range'
        When call TestShufRange
        The output should equal '1'
        The status should be success
    End
End

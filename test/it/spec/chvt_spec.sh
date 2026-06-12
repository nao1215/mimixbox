Describe 'chvt'
    Include console-tools/chvt_test.sh

    It 'rejects a non-numeric VT'
        When call TestChvtBadN
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a VT number'
        When call TestChvtNoN
        The output should equal 'rc=1'
        The status should be success
    End
End

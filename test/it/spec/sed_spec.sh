Describe 'sed'
    Include editors/sed_test.sh

    It 'substitutes the first match'
        When call TestSedSubstitute
        The output should equal 'hello sed'
        The status should be success
    End
    It 'substitutes globally'
        When call TestSedGlobal
        The output should equal 'b b b'
        The status should be success
    End
    It 'deletes a line by number'
        When call TestSedDelete
        The line 1 of output should equal '1'
        The line 2 of output should equal '3'
        The status should be success
    End
    It 'prints a single line with -n'
        When call TestSedPrintN
        The output should equal 'y'
        The status should be success
    End
End

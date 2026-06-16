Describe 'cut --complement on fields'
    Include shellutils/cut_test.sh

    It 'keeps the fields not selected'
        When call TestCutComplementFields
        The output should equal 'a,c'
        The status should be success
    End
End

Describe 'cut --complement on bytes'
    Include shellutils/cut_test.sh

    It 'keeps the bytes not selected'
        When call TestCutComplementBytes
        The output should equal 'ade'
        The status should be success
    End
End

Describe 'cut -z zero-terminated fields'
    Include shellutils/cut_test.sh

    It 'splits and joins records on NUL'
        When call TestCutZeroTerminatedFields
        The output should equal 'b|e|'
        The status should be success
    End
End

Describe 'cut -z zero-terminated bytes'
    Include shellutils/cut_test.sh

    It 'cuts bytes from each NUL-delimited record'
        When call TestCutZeroTerminatedBytes
        The output should equal 'ab|de|'
        The status should be success
    End
End

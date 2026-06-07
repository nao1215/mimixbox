Describe 'cut selects a field'
    Include shellutils/cut_test.sh

    It 'prints the chosen field'
        When call TestCutField
        The output should equal 'b'
        The status should be success
    End
End

Describe 'cut selects multiple fields'
    Include shellutils/cut_test.sh

    It 'prints the chosen fields joined by the delimiter'
        When call TestCutFields
        The output should equal 'a,c'
        The status should be success
    End
End

Describe 'cut selects a field range'
    Include shellutils/cut_test.sh

    It 'prints from the field to the end'
        When call TestCutFieldRange
        The output should equal 'b,c,d'
        The status should be success
    End
End

Describe 'cut selects characters'
    Include shellutils/cut_test.sh

    It 'prints the chosen character range'
        When call TestCutChars
        The output should equal 'abc'
        The status should be success
    End
End

Describe 'cut without a list'
    Include shellutils/cut_test.sh

    It 'reports an error'
        When call TestCutNoList
        The error should equal 'cut: you must specify a list of bytes, characters, or fields'
        The status should be failure
    End
End

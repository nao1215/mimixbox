Describe 'dd copies stdin to stdout'
    Include shellutils/dd_test.sh

    It 'reproduces the input'
        When call TestDdCopy
        The output should equal 'hello world'
        The status should be success
    End
End

Describe 'dd with count'
    Include shellutils/dd_test.sh

    It 'copies only the requested blocks'
        When call TestDdCount
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'dd conv=ucase'
    Include shellutils/dd_test.sh

    It 'upper-cases the data'
        When call TestDdUcase
        The output should equal 'ABC'
        The status should be success
    End
End

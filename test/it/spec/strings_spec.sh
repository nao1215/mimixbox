Describe 'strings from pipe'
    Include textutils/strings_test.sh
    result() { %text
        #|hello
        #|world
    }
    It 'prints printable sequences'
        When call TestStringsPipe
        The output should equal "$(result)"
        The status should be success
    End
End

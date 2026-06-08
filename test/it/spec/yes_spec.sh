Describe 'yes default'
    Include shellutils/yes_test.sh
    result() { %text
        #|y
        #|y
        #|y
    }
    It 'repeats y until the reader closes'
        When call TestYesHead
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'yes with a string'
    Include shellutils/yes_test.sh
    result() { %text
        #|mimix
        #|mimix
    }
    It 'repeats the given string'
        When call TestYesString
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'uniq removes adjacent duplicate lines'
    Include shellutils/uniq_test.sh

    result() { %text
        #|a
        #|b
        #|c
    }

    It 'collapses repeated lines'
        When call TestUniqBasic
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'uniq -c counts occurrences'
    Include shellutils/uniq_test.sh

    result() { %text
        #|      2 a
        #|      1 b
        #|      3 c
    }

    It 'prefixes each line with its count'
        When call TestUniqCount
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'uniq -d prints only repeated lines'
    Include shellutils/uniq_test.sh

    result() { %text
        #|a
        #|c
    }

    It 'prints duplicated lines once'
        When call TestUniqRepeated
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'uniq -u prints only unique lines'
    Include shellutils/uniq_test.sh

    It 'prints lines that never repeat'
        When call TestUniqUnique
        The output should equal 'b'
        The status should be success
    End
End

Describe 'sort lexically'
    Include shellutils/sort_test.sh

    result() { %text
        #|apple
        #|banana
        #|cherry
    }

    It 'sorts lines alphabetically'
        When call TestSortLexical
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort numerically'
    Include shellutils/sort_test.sh

    result() { %text
        #|1
        #|2
        #|10
    }

    It 'sorts by numeric value'
        When call TestSortNumeric
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort in reverse'
    Include shellutils/sort_test.sh

    result() { %text
        #|c
        #|b
        #|a
    }

    It 'reverses the order'
        When call TestSortReverse
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort unique'
    Include shellutils/sort_test.sh

    result() { %text
        #|a
        #|b
    }

    It 'drops duplicate lines'
        When call TestSortUnique
        The output should equal "$(result)"
        The status should be success
    End
End

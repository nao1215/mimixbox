Describe 'seq with one operand'
    Include shellutils/seq_test.sh

    result() { %text
        #|1
        #|2
        #|3
    }

    It 'counts from 1 to LAST'
        When call TestSeqLast
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'seq with first and last'
    Include shellutils/seq_test.sh

    result() { %text
        #|2
        #|3
        #|4
        #|5
    }

    It 'counts from FIRST to LAST'
        When call TestSeqFirstLast
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'seq with increment'
    Include shellutils/seq_test.sh

    result() { %text
        #|1
        #|3
        #|5
        #|7
        #|9
    }

    It 'counts by INCREMENT'
        When call TestSeqFirstIncrementLast
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'seq with a separator'
    Include shellutils/seq_test.sh

    It 'joins the numbers with the separator'
        When call TestSeqSeparator
        The output should equal '1,2,3'
        The status should be success
    End
End

Describe 'seq with equal width'
    Include shellutils/seq_test.sh

    result() { %text
        #|08
        #|09
        #|10
    }

    It 'pads numbers with leading zeros'
        When call TestSeqEqualWidth
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'seq with an invalid operand'
    Include shellutils/seq_test.sh

    It 'reports an error'
        When call TestSeqInvalid
        The error should equal "seq: invalid floating point argument: 'abc'"
        The status should be failure
    End
End

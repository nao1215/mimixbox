Describe 'cal for a given month and year'
    Include shellutils/cal_test.sh

    result() { %text
        #|   November 2023
        #|Su Mo Tu We Th Fr Sa
        #|          1  2  3  4
        #| 5  6  7  8  9 10 11
        #|12 13 14 15 16 17 18
        #|19 20 21 22 23 24 25
        #|26 27 28 29 30
    }

    It 'prints the month calendar'
        When call TestCalMonthYear
        The output should equal "$(result)"
        The status should be success
    End
End

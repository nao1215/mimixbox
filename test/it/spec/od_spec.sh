Describe 'od -c'
    Include shellutils/od_test.sh

    result() { %text
        #|0000000   A   B   C  \n
        #|0000004
    }

    It 'dumps characters with C escapes'
        When call TestOdChar
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'od -A x -t x1'
    Include shellutils/od_test.sh

    result() { %text
        #|000000 41 42
        #|000002
    }

    It 'dumps hex bytes with hex addresses'
        When call TestOdHex
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'od -A n'
    Include shellutils/od_test.sh

    It 'suppresses the address column'
        When call TestOdNoAddr
        The output should equal ' 101'
        The status should be success
    End
End

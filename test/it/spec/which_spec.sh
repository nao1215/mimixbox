Describe 'Search mimixbox'
    Include shellutils/which_test.sh
    It 'print /usr/local/bin/mimixbox'
        When call TestWhichExistBinary
        The output should equal '/usr/local/bin/mimixbox'
        The status should be success
    End
End

Describe 'Search binary that does not exist in system'
    Include shellutils/which_test.sh
    It 'print nothig'
        When call TestWhichNoExistBinary
        The output should equal ''
        The status should be failure
    End
End

Describe 'Search three binary.'
    Include shellutils/which_test.sh

    result() { %text
      #|/usr/local/bin/mimixbox
      #|/usr/local/bin/cat
      #|/usr/local/bin/tac
    }

    It 'print paths of three binary'
        When call TestWhichThreeBinary
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Search three binary. One of three binary does not exist'
    Include shellutils/which_test.sh

    result() { %text
      #|/usr/local/bin/mimixbox
      #|/usr/local/bin/tac
    }

    It 'print paths of two binary and exit-status is fail'
        When call TestWhichOneOfThreeBinNotExist
        The output should equal "$(result)"
        The status should be failure
    End
End

Describe 'Which without no operand'
    Include shellutils/which_test.sh

    It 'print nothing'
        When call TestWhichWithoutOperand
        The output should equal ''
        The status should be failure
    End
End

Describe 'Which data from pipe'
    Include shellutils/which_test.sh

    It 'print nothing'
        When call TestWhichDataFromPipe
        The output should equal ''
        The status should be failure
    End
End
# The expected paths are derived from wherever MimixBox actually resolves on
# PATH, so the suite stays correct under the isolated end-to-end harness
# (test/it/.mbbin) without hardcoding an install prefix like /usr/local/bin.
mb_bindir() { dirname "$(command -v mimixbox)"; }

Describe 'Search mimixbox'
    Include shellutils/which_test.sh
    It 'prints the MimixBox path'
        When call TestWhichExistBinary
        The output should equal "$(mb_bindir)/mimixbox"
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

    result() {
        d=$(mb_bindir)
        printf '%s\n%s\n%s\n' "$d/mimixbox" "$d/cat" "$d/tac"
    }

    It 'print paths of three binary'
        When call TestWhichThreeBinary
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Search three binary. One of three binary does not exist'
    Include shellutils/which_test.sh

    result() {
        d=$(mb_bindir)
        printf '%s\n%s\n' "$d/mimixbox" "$d/tac"
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

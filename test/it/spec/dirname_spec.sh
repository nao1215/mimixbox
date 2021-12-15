Describe 'Get dirname from /home/nao/test.txt'
    Include shellutils/dirname_test.sh

    It 'print /home/nao'
        When call TestDirnameAbsFilePath
        The output should equal "/home/nao"
        The status should be success
    End
End

Describe 'Get dirname without extension'
    Include shellutils/dirname_test.sh

    It 'print test'
        When call TestDirnameFilenameWithoutExt
        The output should equal "/home/nao"
        The status should be success
    End
End

Describe 'Get dirname for hidden directory'
    Include shellutils/dirname_test.sh

    It 'print /home/nao/.test'
        When call TestDirnameHiddenFile
        The output should equal "/home/nao"
        The status should be success
    End
End

Describe 'Get dirname without operand'
    Include shellutils/dirname_test.sh

    It 'print error'
        When call TestDirnameNoOpertand
        The error should equal "dirname: no operand"
        The status should be failure
    End
End

Describe 'Get dirname for root directory'
    Include shellutils/dirname_test.sh

    It 'print /'
        When call TestDirnameRoot
        The output should equal "/"
        The status should be success
    End
End

Describe 'Get dirname for empty string'
    Include shellutils/dirname_test.sh

    It 'print "."'
        When call TestDirnameEmptyString
        The output should equal "."
        The status should be success
    End
End

Describe 'Get dirname for three arguments'
    Include shellutils/dirname_test.sh

    result() { %text
        #|/bin
        #|/home
        #|/
    }

    It 'print "/bin" "/home" "/" with line feed'
        When call TestDirnameWithThreeArg
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get dirname for three arguments with zero option'
    Include shellutils/dirname_test.sh

    It 'print "/bin" "/home" "/" without line feed'
        When call TestDirnameThreeArgWithZeroOption
        The output should equal "/bin/home/"
        The status should be success
    End
End

Describe 'Get dirname with environment variable'
    Include shellutils/dirname_test.sh

    It 'print /aaa/bbb/ccc'
        When call TestDirnameFilenameWithEnvVar
        The output should equal "/aaa/bbb/ccc"
        The status should be success
    End
End
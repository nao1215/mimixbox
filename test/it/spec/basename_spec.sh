Describe 'Get basename from /home/nao/test.txt'
    Include shellutils/basename_test.sh

    It 'show test.txt'
        When call TestBasenameFilenameWithExt
        The output should equal "test.txt"
        The status should be success
    End
End

Describe 'Get basename from /home/nao/test'
    Include shellutils/basename_test.sh

    It 'show test'
        When call TestBasenameFilenameWithoutExt
        The output should equal "test"
        The status should be success
    End
End

Describe 'Get basename from /home/nao/.test'
    Include shellutils/basename_test.sh

    It 'show .test'
        When call TestBasenameHiddenFile
        The output should equal ".test"
        The status should be success
    End
End

Describe 'Get basename from /home/nao/'
    Include shellutils/basename_test.sh

    It 'show nao'
        When call TestBasenameEndsWithThrash
        The output should equal "nao"
        The status should be success
    End
End

Describe 'Get basename without operand'
    Include shellutils/basename_test.sh

    It 'show error'
        When call TestBasenameNoOpertand
        The error should equal "basename: no operand"
        The status should be failure
    End
End

Describe 'Get basename /'
    Include shellutils/basename_test.sh

    It 'show /'
        When call TestBasenameRoot
        The output should equal "/"
        The status should be success
    End
End

Describe 'Get basename empty string'
    Include shellutils/basename_test.sh

    It 'show ""'
        When call TestBasenameEmptyString
        The output should equal ""
        The status should be success
    End
End

Describe 'Get basename with three arguments'
    Include shellutils/basename_test.sh

    It 'show "basename"'
        When call TestBasenameWithThreeArg
        The output should equal "basename"
        The status should be success
    End
End


Describe 'Get basename three arguments with multiple options'
    Include shellutils/basename_test.sh

    result() { %text
        #|basename
        #|nao
        #|home
    }


    It 'show three basename(basename, nao, home)'
        When call TestBasenameThreeArgWithMultipleOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get basename three arguments with multiple/zero options'
    Include shellutils/basename_test.sh

    It 'show three basename(basenamenaohome)'
        When call TestBasenameThreeArgWithMultipleAndZeroOption
        The output should equal "basenamenaohome"
        The status should be success
    End
End

Describe 'Get basename with suffix options'
    Include shellutils/basename_test.sh

    It 'show basename without suffix'
        When call TestBasenameWithSuffixOption
        The output should equal "test"
        The status should be success
    End
End
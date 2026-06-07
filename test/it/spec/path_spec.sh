Describe 'path with --basename'
    Include shellutils/path_test.sh

    It 'prints the base name'
        When call TestPathBasename
        The output should equal 'test.txt'
        The status should be success
    End
End

Describe 'path with --dirname'
    Include shellutils/path_test.sh

    It 'prints the directory'
        When call TestPathDirname
        The output should equal '/home/nao'
        The status should be success
    End
End

Describe 'path with --extension'
    Include shellutils/path_test.sh

    It 'prints the extension'
        When call TestPathExtension
        The output should equal '.txt'
        The status should be success
    End
End

Describe 'path with --canonical'
    Include shellutils/path_test.sh

    It 'prints the cleaned path'
        When call TestPathCanonical
        The output should equal '/home/nao/test.txt'
        The status should be success
    End
End

Describe 'path with no operand'
    Include shellutils/path_test.sh

    It 'reports an error'
        When call TestPathNoOperand
        The error should equal 'path: missing operand'
        The status should be failure
    End
End

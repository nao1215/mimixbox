Describe 'tee writes to stdout'
    Include shellutils/tee_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'echoes standard input to stdout'
        When call TestTeeStdoutAndFile
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'tee writes to a file'
    Include shellutils/tee_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'also writes the input to the file'
        When call TestTeeFileContents
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'tee -a appends to a file'
    Include shellutils/tee_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|one
        #|two
    }

    It 'keeps the existing content'
        When call TestTeeAppend
        The output should equal "$(result)"
        The status should be success
    End
End

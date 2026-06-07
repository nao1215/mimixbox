Describe 'base64 encode from pipe'
    Include shellutils/base64_test.sh

    It 'encodes standard input'
        When call TestBase64EncodePipe
        The output should equal 'aGVsbG8K'
        The status should be success
    End
End

Describe 'base64 encode from file'
    Include shellutils/base64_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'encodes the file contents'
        When call TestBase64EncodeFile
        The output should equal 'aGVsbG8K'
        The status should be success
    End
End

Describe 'base64 decode from pipe'
    Include shellutils/base64_test.sh

    It 'decodes standard input'
        When call TestBase64DecodePipe
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'base64 round trip'
    Include shellutils/base64_test.sh

    It 'returns the original text'
        When call TestBase64RoundTrip
        The output should equal 'MimixBox'
        The status should be success
    End
End

Describe 'base64 with a non-existent file'
    Include shellutils/base64_test.sh

    It 'reports an error'
        When call TestBase64NoExistFile
        The error should equal 'base64: /no_exist_file: no such file or directory'
        The status should be failure
    End
End

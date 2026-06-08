Describe 'base32 encode from pipe'
    Include textutils/base32_test.sh
    It 'encodes standard input'
        When call TestBase32EncodePipe
        The output should equal 'NBSWY3DPBI======'
        The status should be success
    End
End

Describe 'base32 decode from pipe'
    Include textutils/base32_test.sh
    It 'decodes standard input'
        When call TestBase32DecodePipe
        The output should equal 'hello'
        The status should be success
    End
End

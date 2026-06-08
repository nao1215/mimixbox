Describe 'gzip/gunzip roundtrip'
    Include shellutils/gzip_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'compresses and decompresses back to the original'
        When call TestGzipRoundtrip
        The output should equal 'hello gzip roundtrip'
        The status should be success
    End
End

Describe 'compress and uncompress'
    Include archival/compress_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'round-trips a file through compress and uncompress'
        When call TestCompressRoundTrip
        The output should equal 'compress me compress me compress me'
        The status should be success
    End
End

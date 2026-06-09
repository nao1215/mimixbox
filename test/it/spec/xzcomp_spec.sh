Describe 'compression applets'
    Include archival/xzcomp_test.sh

    It 'xz | xzcat round-trips data'
        When call TestXzRoundTrip
        The output should equal 'roundtrip-xz'
        The status should be success
    End
    It 'lzma | unlzma round-trips data'
        When call TestLzmaRoundTrip
        The output should equal 'roundtrip-lzma'
        The status should be success
    End
    It 'zcat decompresses a gzip file to stdout'
        When call TestZcatGzip
        The output should equal 'gz-payload'
        The status should be success
    End
    It 'pipe_progress passes stdin through to stdout'
        When call TestPipeProgress
        The output should equal 'pass-through'
        The status should be success
    End
End

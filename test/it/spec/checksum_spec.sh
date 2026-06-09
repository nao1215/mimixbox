Describe 'checksums'
    Include textutils/checksum_test.sh

    It 'sum prints the BSD checksum and block count'
        When call TestSum
        The output should equal '36979     1'
        The status should be success
    End
    It 'crc32 prints the CRC-32 of stdin'
        When call TestCrc32
        The output should equal '363a3020  -'
        The status should be success
    End
    It 'sha384sum prints the SHA-384 digest'
        When call TestSha384
        The status should be success
        The output should include '1d0f284efe3edea4b9ca3bd514fa134b17eae361ccc7a1eefeff801b9bd6604e'
    End
End

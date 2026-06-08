Describe 'cksum from pipe'
    Include textutils/cksum_test.sh
    It 'prints the CRC checksum and byte count'
        When call TestCksumPipe
        The output should equal '3015617425 6'
        The status should be success
    End
End

Describe 'uuencode / uudecode / usleep'
    Include textutils/checksum_test.sh

    It 'uuencode | uudecode round-trips (traditional)'
        When call TestUuRoundTrip
        The output should equal 'round trip data'
        The status should be success
    End
    It 'uuencode -m | uudecode round-trips (base64)'
        When call TestUuBase64
        The output should equal 'base64 round trip'
        The status should be success
    End
    It 'usleep waits and exits 0'
        When call TestUsleep
        The output should equal 'slept'
        The status should be success
    End
End

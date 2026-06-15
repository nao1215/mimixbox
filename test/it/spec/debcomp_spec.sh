Describe 'bzip2, lzop and Debian package applets'
    Include archival/debcomp_test.sh

    It 'bzip2 | bzip2 -dc round-trips data'
        When call TestBzip2RoundTrip
        The output should equal 'roundtrip-bzip2'
        The status should be success
    End
    It 'lzop | lzopcat round-trips data'
        When call TestLzopRoundTrip
        The output should equal 'roundtrip-lzop'
        The status should be success
    End
    It 'lzop | unlzop -c round-trips data'
        When call TestUnlzopRoundTrip
        The output should equal 'roundtrip-unlzop'
        The status should be success
    End
    It 'dpkg-deb -c lists package contents'
        When call TestDpkgDebContents
        The output should equal 'has-hello'
        The status should be success
    End
    It 'dpkg-deb -f prints a control field'
        When call TestDpkgDebField
        The output should equal 'hello'
        The status should be success
    End
    It 'dpkg -x extracts the data tarball path-safely'
        When call TestDpkgExtract
        The output should equal 'extracted'
        The status should be success
    End
    It 'dpkg rejects unsupported database operations'
        When call TestDpkgUnsupported
        The output should equal 'rejected'
        The status should be success
    End
End

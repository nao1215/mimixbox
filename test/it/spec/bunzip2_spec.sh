Describe 'bunzip2'
    Include archival/bunzip2_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'decompresses a .bz2 file to stdout with -c'
        When call TestBunzip2Stdout
        The output should equal 'bunzip2 payload'
        The status should be success
    End
End

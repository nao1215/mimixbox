Describe 'gunzip'
    Include archival/gunzip_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'decompresses a .gz file to stdout with -c'
        When call TestGunzipStdout
        The output should equal 'gunzip payload'
        The status should be success
    End
End

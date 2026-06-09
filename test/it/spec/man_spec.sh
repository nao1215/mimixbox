Describe 'man'
    Include textutils/man_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'shows a plain manual page'
        When call TestManPlain
        The output should equal 'FOO(1)
the foo page'
        The status should be success
    End
    It 'decompresses a gzipped manual page'
        When call TestManGzip
        The output should equal 'BAR(1)
the bar page'
        The status should be success
    End
    It 'reports a missing page with exit 16'
        When call TestManNotFound
        The output should equal 'exit=16'
        The status should be success
    End
End

Describe 'grep GNU flags'
    Include findutils/grep_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints trailing context with -A1'
        When call TestGrepAfter
        The output should equal 'MATCH
b'
        The status should be success
    End

    It 'prints leading context with -B1'
        When call TestGrepBefore
        The output should equal 'a
MATCH'
        The status should be success
    End

    It 'prints surrounding context with -C1'
        When call TestGrepContext
        The output should equal 'a
MATCH
b'
        The status should be success
    End

    It 'separates non-contiguous groups with --'
        When call TestGrepGroupSeparator
        The line 3 of output should equal '--'
        The status should be success
    End

    It 'searches only included files with --include'
        When call TestGrepInclude
        The output should include 'keep.go'
        The output should not include 'skip.txt'
        The status should be success
    End

    It 'skips excluded files with --exclude'
        When call TestGrepExclude
        The output should not include 'app.log'
        The status should be success
    End

    It 'skips excluded directories with --exclude-dir'
        When call TestGrepExcludeDir
        The output should not include 'vendor'
        The status should be success
    End

    It 'highlights matches with --color=always'
        When call TestGrepColor
        The output should include 'world'
        The status should be success
    End

    It 'prints byte offsets with -b'
        When call TestGrepByteOffset
        The output should equal '4:bbb'
        The status should be success
    End

    It 'prints files without a match with -L'
        When call TestGrepFilesWithoutMatch
        The output should include 'miss.txt'
        The output should not include 'hit.txt'
        The status should be success
    End
End

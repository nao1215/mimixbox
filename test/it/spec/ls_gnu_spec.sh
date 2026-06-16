Describe 'ls GNU presentation flags'
    Include fileutils/ls_gnu_test.sh
    BeforeEach 'GnuSetup'
    AfterEach 'GnuCleanUp'

    # --- #722 color -------------------------------------------------------
    It 'colors directories with --color=always'
        When call TestColorAlways
        The output should include 'adir'
        The output should include "$(printf '\033[01;34m')"
        The status should be success
    End
    It 'emits no escapes with --color=never'
        When call TestColorNever
        The output should not include "$(printf '\033')"
        The status should be success
    End

    # --- #723 indicators --------------------------------------------------
    It 'appends / * @ with -F'
        When call TestClassify
        The output should include 'adir/'
        The output should include 'run.sh*'
        The output should include 'link@'
        The status should be success
    End
    It 'omits * for executables with --file-type'
        When call TestFileType
        The output should include 'adir/'
        The output should not include 'run.sh*'
        The status should be success
    End
    It 'marks only dirs with --indicator-style=slash'
        When call TestIndicatorSlash
        The output should include 'adir/'
        The output should not include 'link@'
        The status should be success
    End

    # --- #724 sorting -----------------------------------------------------
    It 'lists largest first with --sort=size'
        When call TestSortSize
        The line 1 of output should equal 'big.txt'
        The status should be success
    End
    It 'lists directories first with --group-directories-first'
        When call TestGroupDirs
        The line 1 of output should equal 'adir'
        The status should be success
    End

    # --- #725 hide / ignore ----------------------------------------------
    It 'drops matches with --ignore'
        When call TestIgnoreLog
        The output should not include 'a.log'
        The status should be success
    End
    It 'drops matches with --hide'
        When call TestHideTmp
        The output should not include 'tmpfile'
        The status should be success
    End
    It 'keeps hidden matches when -a is given'
        When call TestHideTmpWithAll
        The output should include 'tmpfile'
        The status should be success
    End

    # --- #726 inode / block-size -----------------------------------------
    It 'prints an inode number with -i'
        When call TestInode
        The output should match pattern '[0-9]* *small.txt'
        The status should be success
    End
    It 'scales sizes to 1024-byte blocks with -k'
        When call TestBlockSize
        The output should include ' 5 '
        The status should be success
    End
End

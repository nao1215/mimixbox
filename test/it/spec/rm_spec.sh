Describe 'Remove one file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|2.txt
        #|3.txt
        #|inner
    }
    It 'remove one file.'
        When call TestRmOneFile
        The output should equal "$(result)"
    End
End

Describe 'Check status after removing one file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'remove one file.'
        When call TestRmOneStatus
        The status should be success
    End
End

Describe 'Remove file using wildcard'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'remove all text file.'
        When call TestRmFileWithWildcard
        The output should equal "inner"
    End
End

Describe 'Check status after removing file using wildcard'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestRmFileWithWildcardStatus
        The status should be success
    End
End

Describe 'Remove three file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'remove three text file.'
        When call TestRmThreeFileAtSameTime
        The output should equal "inner"
    End
End

Describe 'Check status after removing three file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestRmThreeFileAtSameTimeStatus
        The status should be success
    End
End

Describe 'Remove three file. One of them does not exist'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|2.txt
        #|inner
    }

    It 'remove two text file and print error.'
        When call TestRmThreeFileWithNoExistFile
        The output should equal "$(result)"
        The error should equal "rm: can't remove /tmp/mimixbox/it/rm/no_exist_file.txt: No such file or directory exists"
    End
End

Describe 'Check status after removing three file. One of them does not exist'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status failure'
        When call TestRmThreeFileWithNoExistFileStatus
        The error should equal "rm: can't remove /tmp/mimixbox/it/rm/no_exist_file.txt: No such file or directory exists"
        The status should be failure
    End
End

Describe 'Remove directory without recursive options'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
    }

    It 'can not remove directory'
        When call TestRmDirWithoutRecursiveOption
        The error should equal "rm: can't remove /tmp/mimixbox/it/rm: It's directory"
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check status after removing directory without recursive option'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status failure'
        When call TestRmDirWithoutRecursiveOptionStatus
        The error should equal "rm: can't remove /tmp/mimixbox/it/rm: It's directory"
        The status should be failure
    End
End

Describe 'Remove directory with recursive options'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'can remove directory'
        When call TestRmDirWithRecursiveOption
        The output should equal ""
        The status should be success
    End
End

Describe 'Check status removing directory with recursive options'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestRmDirWithRecursiveOptionStatus
        The status should be success
    End
End

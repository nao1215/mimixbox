Describe 'Rename file-name'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'rename 1.txt to rename.txt'
        When call TestMvRename
        The output should equal '/tmp/mimixbox/it/mv/rename.txt'
    End
End

Describe 'Check status after renaming file-name'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMvRenameStatus
        The status should be success
    End
End

Describe 'Move file to inner direcroty'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|inner.txt
    }

    It 'move a file to inner directory'
        When call TestMvMoveFile
        The output should equal "$(result)"
    End
End

Describe 'Check status after moving file'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMvMoveFileStatus
        The status should be success
    End
End

Describe 'Move three file to inner direcroty'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner.txt
    }

    It 'move three file to inner directory'
        When call TestMvThreeFileAtSameTime
        The output should equal "$(result)"
    End
End

Describe 'Check status after moving three file'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMvThreeFileAtSameTimeStatus
        The status should be success
    End
End

Describe 'Move three file to inner direcroty. And One of three can not move'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|3.txt
        #|inner.txt
    }

    It 'move two file to inner directory'
        When call TestMvThreeFileAndOneOfThreeFail
        The error should equal "mv: /tmp/mimixbox/it/mv/no_exist_file doesn't exist"
        The output should equal "$(result)"
    End
End

Describe 'Check status after moving three files and one of them fail'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status failure'
        When call TestMvThreeFileAndOneOfThreeFailStatus
        The error should equal "mv: /tmp/mimixbox/it/mv/no_exist_file doesn't exist"
        The status should be failure
    End
End

Describe 'Move directory to directrory'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
        #|mv2
    }

    It 'move two file to inner directory'
        When call TestMvDirToDir
        The output should equal "$(result)"
    End
End

Describe 'Check status after moving directory'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMvDirToDirStatus
        The status should be success
    End
End

Describe 'Move three directory'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
        #|mv2
        #|mv3
        #|mv4
    }

    It 'move three directory under /tmp/mimixbox/it/mv'
        When call TestMvThreeDirs
        The output should equal "$(result)"
    End
End

Describe 'Check status after moving three directories'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMvThreeStatus
        The status should be success
    End
End

Describe 'Move three directories to inner direcroty. And One of three can not move'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|inner.txt
        #|mv2
        #|mv4
    }

    It 'move two directory to inner directory'
        When call TestMvThreeDirsAndOneOfThreeFail
        The error should equal "mv: /tmp/mimixbox/it/mv/no_exist_dir doesn't exist"
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check status after moving directory'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status failure'
        When call TestMvThreeDirsAndOneOfThreeFailStatus
        The error should equal "mv: /tmp/mimixbox/it/mv/no_exist_dir doesn't exist"
        The status should be failure
    End
End

Describe 'Move file with back up option.'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status failure'
        When call TestMvFileAtSampePath
        The error should equal "mv: source '/tmp/mimixbox/it/mv/1.txt' and destination '/tmp/mimixbox/it/mv/1.txt' is same"
        The status should be failure
    End
End

Describe 'Move file. Source filename and dest file name is same'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
        #|inner.txt
    }

    It 'overwrite file'
        When call TestMvSrcAndDestIsSameName
        The output should equal "$(result)"
    End
End

Describe 'Check status. Source filename and dest file name is same'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestMvSrcAndDestIsSameNameStatus
        The status should be success
    End
End

Describe 'Move file with backup option. Source filename and dest file name is same'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|inner.txt
        #|inner.txt~
    }

    It 'overwrite file'
        When call TestMvSrcAndDestIsSameNameWithBackupOpt
        The output should equal "$(result)"
    End
End

Describe 'Check status. Move file with backup option. Source filename and dest file name is same'
    Include fileutils/mv_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestMvSrcAndDestIsSameNameWithBackupOptStatus
        The status should be success
    End
End
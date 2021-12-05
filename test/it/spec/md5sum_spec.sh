Describe 'Get md5sum of one file'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says d0d8ffef81b3c7160ac655d5939548c5'
        When call TestMd5sumOneFile
        The output should equal 'd0d8ffef81b3c7160ac655d5939548c5  /tmp/mimixbox/it/md5sum/1.txt'
        The status should be success
    End
End

Describe 'Can not get md5sum of one directory'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "It is directory"'
        When call TestMd5sumOneDirectory
        The error should equal 'md5sum: /tmp/mimixbox/it/md5sum: It is directory'
        The status should be failure
    End
End

Describe 'Can not get md5sum of not exist file'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "No such file or directory"'
        When call TestMd5sumNotExistFile
        The error should equal 'md5sum: /not_exist_file: No such file or directory'
        The status should be failure
    End
End

Describe 'Get md5sum of three files'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|d0d8ffef81b3c7160ac655d5939548c5  /tmp/mimixbox/it/md5sum/1.txt
        #|07e280ad4bd77b9321f0ce3386775019  /tmp/mimixbox/it/md5sum/2.txt
        #|15e924f84517598e828f49dc85765bc5  /tmp/mimixbox/it/md5sum/3.txt
    }

    It 'show checksum of three file'
        When call TestMd5sumThreeFiles
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check md5sum with --check option'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|/tmp/mimixbox/it/md5sum/1.txt: OK
        #|/tmp/mimixbox/it/md5sum/2.txt: OK
        #|/tmp/mimixbox/it/md5sum/3.txt: OK
    }

    It 'show all files "OK"'
        When call TestMd5sumWithCheckOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get md5sum for pipe data'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'show checksum'
        When call TestMd5sumDataFromPipe
        The output should equal "d8e8fca2dc0f896fd7cb4cb0031ba249  -"
        The status should be success
    End
End

Describe 'Get md5sum for pipe data and file at same time'
    Include textutils/md5sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'only show checksum of file.'
        When call TestMd5sumFileAndDataFromPipeAtSameTime
        The output should equal "d0d8ffef81b3c7160ac655d5939548c5  /tmp/mimixbox/it/md5sum/1.txt"
        The status should be success
    End
End
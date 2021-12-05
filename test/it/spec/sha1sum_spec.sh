Describe 'Get sha1sum of one file'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says 9dc2936d38932f9ffc6738cb677e4a8722116070'
        When call TestMd5sumOneFile
        The output should equal '9dc2936d38932f9ffc6738cb677e4a8722116070  /tmp/mimixbox/it/sha1sum/1.txt'
        The status should be success
    End
End

Describe 'Can not get sha1sum of one directory'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "It is directory"'
        When call TestMd5sumOneDirectory
        The error should equal 'sha1sum: /tmp/mimixbox/it/sha1sum: It is directory'
        The status should be failure
    End
End

Describe 'Can not get sha1sum of not exist file'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "No such file or directory"'
        When call TestMd5sumNotExistFile
        The error should equal 'sha1sum: /not_exist_file: No such file or directory'
        The status should be failure
    End
End

Describe 'Get sha1sum of three files'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|9dc2936d38932f9ffc6738cb677e4a8722116070  /tmp/mimixbox/it/sha1sum/1.txt
        #|317e30648976d62fae4662fe4435e6568648e8a7  /tmp/mimixbox/it/sha1sum/2.txt
        #|d4e9619d949de0c0182a09757346ad22e80114b3  /tmp/mimixbox/it/sha1sum/3.txt
    }

    It 'show checksum of three file'
        When call TestMd5sumThreeFiles
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check sha1sum with --check option'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|/tmp/mimixbox/it/sha1sum/1.txt: OK
        #|/tmp/mimixbox/it/sha1sum/2.txt: OK
        #|/tmp/mimixbox/it/sha1sum/3.txt: OK
    }

    It 'show all files "OK"'
        When call TestMd5sumWithCheckOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get sha1sum for pipe data'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'show checksum'
        When call TestMd5sumDataFromPipe
        The output should equal "4e1243bd22c66e76c2ba9eddc1f91394e57f9f83  -"
        The status should be success
    End
End

Describe 'Get sha1sum for pipe data and file at same time'
    Include textutils/sha1sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'only show checksum of file.'
        When call TestMd5sumFileAndDataFromPipeAtSameTime
        The output should equal "9dc2936d38932f9ffc6738cb677e4a8722116070  /tmp/mimixbox/it/sha1sum/1.txt"
        The status should be success
    End
End
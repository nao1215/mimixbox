Describe 'Get sha256sum of one file'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says 5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7'
        When call TestMd5sumOneFile
        The output should equal '5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  /tmp/mimixbox/it/sha256sum/1.txt'
        The status should be success
    End
End

Describe 'Can not get sha256sum of one directory'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "It is directory"'
        When call TestMd5sumOneDirectory
        The error should equal 'sha256sum: /tmp/mimixbox/it/sha256sum: It is directory'
        The status should be failure
    End
End

Describe 'Can not get sha256sum of not exist file'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "No such file or directory"'
        When call TestMd5sumNotExistFile
        The error should equal 'sha256sum: /not_exist_file: No such file or directory'
        The status should be failure
    End
End

Describe 'Get sha256sum of three files'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  /tmp/mimixbox/it/sha256sum/1.txt
        #|833d8136112b60552a0f83165a2ebffeac4b0c0249480d651ea58b9073ec925b  /tmp/mimixbox/it/sha256sum/2.txt
        #|8e774f75a5a23c83e6f7d5e92863a2615e0335e06aec18d9c3ec1c5315d1a777  /tmp/mimixbox/it/sha256sum/3.txt
    }

    It 'show checksum of three file'
        When call TestMd5sumThreeFiles
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check sha256sum with --check option'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|/tmp/mimixbox/it/sha256sum/1.txt: OK
        #|/tmp/mimixbox/it/sha256sum/2.txt: OK
        #|/tmp/mimixbox/it/sha256sum/3.txt: OK
    }

    It 'show all files "OK"'
        When call TestMd5sumWithCheckOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get sha256sum for pipe data'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'show checksum'
        When call TestMd5sumDataFromPipe
        The output should equal "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2  -"
        The status should be success
    End
End

Describe 'Get sha256sum for pipe data and file at same time'
    Include textutils/sha256sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'only show checksum of file.'
        When call TestMd5sumFileAndDataFromPipeAtSameTime
        The output should equal "5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  /tmp/mimixbox/it/sha256sum/1.txt"
        The status should be success
    End
End
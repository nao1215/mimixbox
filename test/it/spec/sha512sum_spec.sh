Describe 'Get sha512sum of one file'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says 05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612'
        When call TestMd5sumOneFile
        The output should equal '05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  /tmp/mimixbox/it/sha512sum/1.txt'
        The status should be success
    End
End

Describe 'Can not get sha512sum of one directory'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "It is directory"'
        When call TestMd5sumOneDirectory
        The error should equal 'sha512sum: /tmp/mimixbox/it/sha512sum: It is directory'
        The status should be failure
    End
End

Describe 'Can not get sha512sum of not exist file'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says "No such file or directory"'
        When call TestMd5sumNotExistFile
        The error should equal 'sha512sum: /not_exist_file: No such file or directory'
        The status should be failure
    End
End

Describe 'Get sha512sum of three files'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  /tmp/mimixbox/it/sha512sum/1.txt
        #|cb2389a103184f607973b1acd073dc15310c8172b03f340a52bdc3843621cf9fbc6263c7dbbd786ceb0244f5147a83aa32ce09a485f544093b7fc5c7533e564f  /tmp/mimixbox/it/sha512sum/2.txt
        #|3dafa5f1ec7f09cbe551dc0d4bdb153dedb81104b7e930b7c20733965f7ebb86ee2abea64b6bfa1c54045032865044a3feca5dcc89c28def410b2954094a1890  /tmp/mimixbox/it/sha512sum/3.txt
    }

    It 'show checksum of three file'
        When call TestMd5sumThreeFiles
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Check sha512sum with --check option'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|/tmp/mimixbox/it/sha512sum/1.txt: OK
        #|/tmp/mimixbox/it/sha512sum/2.txt: OK
        #|/tmp/mimixbox/it/sha512sum/3.txt: OK
    }

    It 'show all files "OK"'
        When call TestMd5sumWithCheckOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Get sha512sum for pipe data'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'show checksum'
        When call TestMd5sumDataFromPipe
        The output should equal "0e3e75234abc68f4378a86b3f4b32a198ba301845b0cd6e50106e874345700cc6663a86c1ea125dc5e92be17c98f9a0f85ca9d5f595db2012f7cc3571945c123  -"
        The status should be success
    End
End

Describe 'Get sha512sum for pipe data and file at same time'
    Include textutils/sha512sum_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'only show checksum of file.'
        When call TestMd5sumFileAndDataFromPipeAtSameTime
        The output should equal "05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  /tmp/mimixbox/it/sha512sum/1.txt"
        The status should be success
    End
End
Describe 'Make one named pipe'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'make named pipe by prw-rw-r--'
        When call TestMkfifoOneFifo
        The output should equal 'prw-rw-r--'
    End
End

Describe 'Check status after making one named pipe'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMkfifoOneFileStatus
        The status should be success
    End
End

Describe 'Make three named pipe'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|prw-rw-r--
        #|prw-rw-r--
        #|prw-rw-r--
    }

    It 'make three named pipe by prw-rw-r--'
        When call TestMkfifoThreeFifo
        The output should equal "$(result)"
    End
End

Describe 'Check status after making three named pipe'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestMkfifoThreeFifoStatus
        The status should be success
    End
End

Describe 'Make named pipe at no exist path'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'print "no such file or direcroty"'
        When call TestMkfifoNoExistPath
        The error should equal 'mkfifo: /no_exist_path/fifo: no such file or directory'
        The status should be failure
    End
End

Describe 'If the same name file already exists, '
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'print "already exist"'
        When call TestMkfifoAlreadyExistSameName
        The error should equal "mkfifo: can't make /tmp/mimixbox/it/mkfifo/1: already exist"
        The status should be failure
    End
End

Describe 'If mkfifo fail to create one file while creating three files,'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1
        #|3
    }

    It 'make two named pipe'
        When call TestMkfifoThreeFileAndCreateOneFileFailed
        The output should equal "$(result)"
        The error should equal 'mkfifo: /no_exist_path/fifo: no such file or directory'
        The status should be success
    End
End

Describe 'Check status If mkfifo fail to create one while creating three files,'
    Include fileutils/mkfifo_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status failure'
        When call TestMkfifoThreeFileAndCreateOneFileFailedStatus
        The error should equal 'mkfifo: /no_exist_path/fifo: no such file or directory'
        The status should be failure
    End
End

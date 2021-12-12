Describe 'unix2dos CRLF file'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/1.txt to DOS format...
        #|/tmp/mimixbox/it/unix2dos/1.txt: ASCII text, with CRLF line terminators
    }

    It 'change LF'
        When call TestUnix2dosCRLF
        The output should equal "$(result)"
    End
End

Describe 'Check status after unix2dos CRLF file'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/1.txt to DOS format...
    }

    It 'status success'
        When call TestUnix2dosCRLFStatus
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'unix2dos three CRLF file'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/1.txt to DOS format...
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/2.txt to DOS format...
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/3.txt to DOS format...
        #|/tmp/mimixbox/it/unix2dos/1.txt: ASCII text, with CRLF line terminators
        #|/tmp/mimixbox/it/unix2dos/2.txt: ASCII text, with CRLF line terminators
        #|/tmp/mimixbox/it/unix2dos/3.txt: ASCII text, with CRLF line terminators
    }

    It 'change LF'
        When call TestUnix2dosThreeFileAtSameTime
        The output should equal "$(result)"
    End
End

Describe 'Check status after unix2dos three CRLF file'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/1.txt to DOS format...
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/2.txt to DOS format...
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/3.txt to DOS format...
    }

    It 'status success'
        When call TestUnix2dosThreeFileAtSameTimeStatus
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Unix2dos directory'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|unix2dos: skip /tmp/mimixbox/it/unix2dos: not regular file
    }

    It 'print "no regular file error"'
        When call TestUnix2dosDir
        The error should equal "$(result)"
        The status should be failure
    End
End

Describe 'Unix2dos two file and one directory'
    Include textutils/unix2dos_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    output_result() { %text
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/1.txt to DOS format...
        #|unix2dos: converting file /tmp/mimixbox/it/unix2dos/3.txt to DOS format...
    }

    error_result() { %text
        #|unix2dos: skip /tmp/mimixbox/it/unix2dos: not regular file
    }


    It 'status error'
        When call TestUnix2dosOneOfThreeFail
        The output should equal "$(output_result)"
        The error should equal "$(error_result)"
        The status should be failure
    End
End

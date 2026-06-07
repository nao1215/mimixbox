Describe 'head default'
    Include textutils/head_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|1
        #|2
        #|3
        #|4
        #|5
        #|6
        #|7
        #|8
        #|9
        #|10
    }

    It 'prints the first 10 lines'
        When call TestHeadDefault
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'head with --lines'
    Include textutils/head_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|1
        #|2
        #|3
    }

    It 'prints the first N lines'
        When call TestHeadLines
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'head with --bytes'
    Include textutils/head_test.sh

    It 'prints the first N bytes'
        When call TestHeadBytes
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'head from pipe'
    Include textutils/head_test.sh

    result() { %text
        #|a
        #|b
    }

    It 'prints the first N lines of stdin'
        When call TestHeadPipe
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'head with a non-existent file'
    Include textutils/head_test.sh

    It 'reports an error'
        When call TestHeadNoExistFile
        The error should equal 'head: /no_exist_file: no such file or directory'
        The status should be failure
    End
End

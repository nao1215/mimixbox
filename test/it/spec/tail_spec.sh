Describe 'tail default'
    Include textutils/tail_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|3
        #|4
        #|5
        #|6
        #|7
        #|8
        #|9
        #|10
        #|11
        #|12
    }

    It 'prints the last 10 lines'
        When call TestTailDefault
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'tail with --lines'
    Include textutils/tail_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|10
        #|11
        #|12
    }

    It 'prints the last N lines'
        When call TestTailLines
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'tail with --bytes'
    Include textutils/tail_test.sh

    It 'prints the last N bytes'
        When call TestTailBytes
        The output should equal 'world'
        The status should be success
    End
End

Describe 'tail from pipe'
    Include textutils/tail_test.sh

    result() { %text
        #|c
        #|d
    }

    It 'prints the last N lines of stdin'
        When call TestTailPipe
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'tail with a non-existent file'
    Include textutils/tail_test.sh

    It 'reports an error'
        When call TestTailNoExistFile
        The error should equal 'tail: /no_exist_file: no such file or directory'
        The status should be failure
    End
End

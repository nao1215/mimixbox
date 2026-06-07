Describe 'tac from file'
    Include textutils/tac_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|third
        #|second
        #|first
    }

    It 'prints the lines in reverse order'
        When call TestTacFile
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'tac from pipe'
    Include textutils/tac_test.sh

    result() { %text
        #|c
        #|b
        #|a
    }

    It 'reverses standard input'
        When call TestTacPipe
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'tac with a non-existent file'
    Include textutils/tac_test.sh

    It 'reports an error'
        When call TestTacNoExistFile
        The error should equal 'tac: /no_exist_file: no such file or directory'
        The status should be failure
    End
End

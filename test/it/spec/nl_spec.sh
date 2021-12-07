Describe 'Show text file with number line'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  sh
        #|     2  ash
        #|     3  csh
        #|     4  bash
    }
    It 'show shell family-name wiht number line'
        When call TestNlNoArg
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Show text file with number line'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  /tmp/mimixbox/it/nl.txt
    }
    It 'show shell family-name wiht number line'
        When call TestNlFromPipeData
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'If pass filename and pipe-data'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  sh
        #|     2  ash
        #|     3  csh
        #|     4  bash
    }
    It 'show only file-data'
        When call TestNlOnlyOperandWithPipeData
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Concatenate two file'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  sh
        #|     2  ash
        #|     3  csh
        #|     4  bash
        #|     5  fish
        #|     6  zsh
    }
    It 'show only file-data'
        When call TestNlConcatenateTwoFile
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Nl command using heaedoc'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  fish
        #|     2  zsh
        #|     3  sh
        #|     4  ash
        #|     5  csh
        #|     6  bash
    }
    It 'show shell family-name wiht number line'
        When call TestNlHeredoc
        The output should equal "$(result)"
    End
End

Describe 'Check status after  using heaedoc'
    Include textutils/nl_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|     1  fish
        #|     2  zsh
        #|     3  sh
        #|     4  ash
        #|     5  csh
        #|     6  bash
    }
    It 'show success'
        When call TestNlHeredoc
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'nl does not exist file.'
    Include textutils/nl_test.sh

    It 'show error'
        When call TestNlNoOperand
        The error should equal "nl: open no_exist_file: no such file or directory"
        The status should be failure
    End
End
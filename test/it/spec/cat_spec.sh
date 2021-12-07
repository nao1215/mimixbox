Describe 'cat file without options'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|sh
        #|ash
        #|csh
        #|bash
    }

    It 'show shell family name'
        When call TestCatNoArg
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat file with --number option'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|     1  sh
        #|     2  ash
        #|     3  csh
        #|     4  bash
    }

    It 'show shell family name with number line.'
        When call TestCatWithNumbetOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat from pipe data'
    Include textutils/cat_test.sh

    It 'show /tmp/mimixbox/it/cat.txt'
        When call TestCatFromPipeData
        The output should equal "/tmp/mimixbox/it/cat.txt"
        The status should be success
    End
End

Describe 'cat only file, not pipe data'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|sh
        #|ash
        #|csh
        #|bash
    }

    It 'show shell family name'
        When call TestCatOnlyOperandWithPipeData
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat two file'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|sh
        #|ash
        #|csh
        #|bash
        #|fish
        #|zsh
    }

    It 'concatenate two file'
        When call TestCatConcatenateTwoFile
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat two file with --number option'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|     1  sh
        #|     2  ash
        #|     3  csh
        #|     4  bash
        #|     5  fish
        #|     6  zsh
    }

    It 'concatenate two file'
        When call TestCatConcatenateTwoFileWithNumberOption
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat heardoc and redirect'
    Include textutils/cat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|fish
        #|zsh
        #|sh
        #|ash
        #|csh
        #|bash
    }

    It 'concatenate file and heardoc'
        When call TestCatHeredoc
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat does not exist file.'
    Include textutils/cat_test.sh

    It 'show error'
        When call TestCatNoOperand
        The error should equal "cat: open no_exist_file: no such file or directory"
        The status should be failure
    End
End
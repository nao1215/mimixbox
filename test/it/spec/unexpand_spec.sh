Describe 'unexpand from pipe'
    Include textutils/unexpand_test.sh

    It 'converts leading spaces to a tab'
        When call TestUnexpandPipe
        The output should equal "$(printf '\ta')"
        The status should be success
    End
End

Describe 'unexpand with --all'
    Include textutils/unexpand_test.sh

    It 'converts internal space runs to tabs'
        When call TestUnexpandAll
        The output should equal "$(printf 'a\t b')"
        The status should be success
    End
End

Describe 'unexpand with a non-existent file'
    Include textutils/unexpand_test.sh

    It 'reports an error'
        When call TestUnexpandNoExistFile
        The error should equal 'unexpand: /no_exist_file: no such file or directory'
        The status should be failure
    End
End

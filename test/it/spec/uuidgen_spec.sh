Describe 'uuidgen'
    Include shellutils/uuidgen_test.sh
    It 'prints a well-formed UUID'
        When call TestUuidgen
        The output should equal 'ok'
        The status should be success
    End
End

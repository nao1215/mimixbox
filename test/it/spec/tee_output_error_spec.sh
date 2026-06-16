Describe 'tee --output-error writable path'
    Include shellutils/tee_output_error_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'copies input and succeeds with an explicit MODE'
        When call TestTeeOutputErrorWritable
        The output should equal 'hello'
        The status should be success
    End
End

Describe 'tee --output-error=warn keeps writing on error'
    Include shellutils/tee_output_error_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'still writes the good file but exits nonzero'
        When call TestTeeOutputErrorWarnContinues
        The output should equal 'payload'
        The status should be failure
    End
End

Describe 'tee --output-error=exit stops at first error'
    Include shellutils/tee_output_error_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'does not create the later good file and exits nonzero'
        When call TestTeeOutputErrorExitStops
        The output should equal 'absent'
        The status should be failure
    End
End

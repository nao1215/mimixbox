Describe 'ed'
    Include editors/ed_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints the buffer with size'
        When call TestEdPrint
        The output should equal '14
one
two
three'
        The status should be success
    End
    It 'appends a line and writes it'
        When call TestEdAppendWrite
        The output should equal 'one
two
INSERTED
three'
        The status should be success
    End
    It 'substitutes text on a line'
        When call TestEdSubstitute
        The output should equal 'one
TWO
three'
        The status should be success
    End
End

Describe 'test string equality'
    Include shellutils/test_test.sh

    It 'is true for equal strings'
        When call TestTestStringEqual
        The output should equal 'rc=0'
        The status should be success
    End
End

Describe 'test integer comparison'
    Include shellutils/test_test.sh

    It 'is true when 2 > 1'
        When call TestTestIntCompare
        The output should equal 'rc=0'
        The status should be success
    End

    It 'is false when 1 > 2'
        When call TestTestIntFalse
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'test file existence'
    Include shellutils/test_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'is true for an existing file'
        When call TestTestFileExists
        The output should equal 'rc=0'
        The status should be success
    End
End

Describe 'test negation'
    Include shellutils/test_test.sh

    It 'negates the expression'
        When call TestTestNegate
        The output should equal 'rc=0'
        The status should be success
    End
End

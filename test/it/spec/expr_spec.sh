Describe 'expr addition'
    Include shellutils/expr_test.sh

    It 'adds two numbers'
        When call TestExprAdd
        The output should equal '13'
        The status should be success
    End
End

Describe 'expr multiplication'
    Include shellutils/expr_test.sh

    It 'multiplies two numbers'
        When call TestExprMul
        The output should equal '12'
        The status should be success
    End
End

Describe 'expr grouping'
    Include shellutils/expr_test.sh

    It 'respects parentheses'
        When call TestExprGroup
        The output should equal '9'
        The status should be success
    End
End

Describe 'expr length'
    Include shellutils/expr_test.sh

    It 'prints the string length'
        When call TestExprLength
        The output should equal '4'
        The status should be success
    End
End

Describe 'expr with a zero result'
    Include shellutils/expr_test.sh

    It 'prints 0 and exits non-zero'
        When call TestExprFalse
        The output should equal '0'
        The status should be failure
    End
End

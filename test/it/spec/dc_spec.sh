Describe 'dc'
    Include shellutils/dc_test.sh

    It 'performs integer division'
        When call TestDcDivide
        The output should equal '2'
        The status should be success
    End
    It 'honors the precision register'
        When call TestDcScale
        The output should equal '2.33'
        The status should be success
    End
    It 'evaluates -e expressions'
        When call TestDcExpr
        The output should equal '1024'
        The status should be success
    End
    It 'stores and loads registers'
        When call TestDcRegisters
        The output should equal '8'
        The status should be success
    End
End

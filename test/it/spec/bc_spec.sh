Describe 'bc'
    Include shellutils/bc_test.sh

    It 'respects operator precedence'
        When call TestBcPrecedence
        The output should equal '14'
        The status should be success
    End
    It 'honors scale for division'
        When call TestBcScale
        The output should equal '2.33'
        The status should be success
    End
    It 'supports variables'
        When call TestBcVars
        The output should equal '25'
        The status should be success
    End
    It 'computes powers'
        When call TestBcPower
        The output should equal '1024'
        The status should be success
    End
End

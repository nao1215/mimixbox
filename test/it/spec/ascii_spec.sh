Describe 'ascii'
    Include console-tools/ascii_test.sh

    It 'prints 128 entries'
        When call TestAsciiLineCount
        The output should equal '128'
        The status should be success
    End
    It 'maps code 65 to A'
        When call TestAsciiCapitalA
        The output should equal '1'
        The status should be success
    End
End

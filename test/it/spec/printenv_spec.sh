Describe 'printenv'
    Include shellutils/printenv_test.sh
    It 'prints an environment variable'
        When call TestPrintenv
        The output should equal 'hello'
        The status should be success
    End
End

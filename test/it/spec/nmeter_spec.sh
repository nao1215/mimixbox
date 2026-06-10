Describe 'nmeter'
    Include procps/nmeter_test.sh

    It 'expands a literal percent and copies text'
        When call TestNmeterLiteral
        The output should equal 'hello % world'
        The status should be success
    End
    It 'expands the total-memory directive'
        When call TestNmeterMem
        The output should equal '1'
        The status should be success
    End
End

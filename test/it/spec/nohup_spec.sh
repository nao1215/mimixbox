Describe 'nohup'
    Include shellutils/nohup_test.sh
    It 'runs the command and passes output through'
        When call TestNohupRuns
        The output should equal 'hello'
        The status should be success
    End
End

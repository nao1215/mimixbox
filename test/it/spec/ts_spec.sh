Describe 'ts'
    Include console-tools/ts_test.sh

    It 'prefixes each line with a timestamp'
        When call TestTs
        The output should equal '2'
        The status should be success
    End
End

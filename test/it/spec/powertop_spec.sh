Describe 'powertop'
    Include procps/powertop_test.sh

    It 'runs and exits zero'
        When call TestPowertopRuns
        The output should equal 'rc=0'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestPowertopHelp
        The status should be success
        The output should include 'Usage: powertop'
        The output should include 'power'
    End
End

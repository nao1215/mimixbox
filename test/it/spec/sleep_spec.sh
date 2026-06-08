Describe 'sleep'
    Include shellutils/sleep_test.sh
    It 'sleeps then returns'
        When call TestSleep
        The output should equal 'slept'
        The status should be success
    End
End

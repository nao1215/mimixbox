Describe 'pwscore'
    Include securityutils/pwscore_test.sh
    It 'scores a common password as zero'
        When call TestPwscoreCommon
        The output should include 'Score: 0/100'
        The status should be success
    End
End

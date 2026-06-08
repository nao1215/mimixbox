Describe 'pwcrack'
    Include securityutils/pwcrack_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'finds a weak password in the wordlist'
        When call TestPwcrack
        The output should include ': secret'
        The status should be success
    End
End

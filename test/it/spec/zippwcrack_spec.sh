Describe 'zip-pwcrack'
    Include securityutils/zippwcrack_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'recovers the ZIP password from the wordlist'
        When call TestZipPwcrack
        The output should include 'password found: hunter2'
        The status should be success
    End
End

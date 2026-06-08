Describe 'ar'
    Include archival/ar_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'lists members'
        When call TestArList
        The line 1 of output should equal 'a.txt'
        The line 2 of output should equal 'b.txt'
        The status should be success
    End

    It 'extracts a member'
        When call TestArExtract
        The output should equal 'alpha'
        The status should be success
    End
End

Describe 'tar'
    Include archival/tar_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates and extracts an archive'
        When call TestTarRoundTrip
        The output should equal 'alpha'
        The status should be success
    End

    It 'lists archive contents'
        When call TestTarList
        The output should include 'src/a.txt'
        The status should be success
    End
End

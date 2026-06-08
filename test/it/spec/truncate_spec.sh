Describe 'truncate'
    Include fileutils/truncate_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'sets the file to the given size'
        When call TestTruncate
        The output should equal '7'
        The status should be success
    End
End

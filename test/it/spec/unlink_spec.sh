Describe 'unlink'
    Include fileutils/unlink_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'removes a single file'
        When call TestUnlink
        The output should equal 'gone'
        The status should be success
    End
End

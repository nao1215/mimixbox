Describe 'link'
    Include fileutils/link_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'creates a hard link sharing contents'
        When call TestLink
        The output should equal 'data'
        The status should be success
    End
End

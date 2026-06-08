Describe 'shred'
    Include fileutils/shred_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'overwrites and removes the file'
        When call TestShredRemove
        The output should equal 'gone'
        The status should be success
    End
End

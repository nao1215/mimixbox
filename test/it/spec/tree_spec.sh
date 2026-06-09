Describe 'tree / nice'
    Include shellutils/tree_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'tree counts directories and files in its summary'
        When call TestTreeSummary
        The output should equal '2 directories, 2 files'
        The status should be success
    End
    It 'tree exits successfully on a readable directory'
        When call TestTreeStatus
        The output should equal '0'
        The status should be success
    End
    It 'nice prints a numeric niceness'
        When call TestNicePrints
        The status should be success
        The output should match pattern '*[0-9]*'
    End
End

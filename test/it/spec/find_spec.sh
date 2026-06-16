Describe 'find'
    Include findutils/find_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'finds a file by -name'
        When call TestFindName
        The output should equal "${MIMIXBOX_IT_ROOT}/find/a.txt"
        The status should be success
    End

    It 'lists directories with -type d'
        When call TestFindTypeDirCount
        The output should equal '2'
        The status should be success
    End

    It 'rejects an unknown predicate'
        When call TestFindUnknown
        The status should be failure
        The error should include 'unknown predicate'
    End

    It 'prints usage for --help'
        When call TestFindHelp
        The output should include 'Usage: find'
        The status should be success
    End

    It 'lists an Options block with --help and --version in --help'
        When call TestFindHelp
        The output should include 'Options:'
        The output should include '--help'
        The output should include '--version'
        The status should be success
    End

    It 'documents the supported subset tokens in --help'
        When call TestFindHelp
        The output should include '-name'
        The output should include '-type'
        The output should include '-print0'
        The output should include '-maxdepth'
        The output should include '-mindepth'
        The status should be success
    End

    It 'prints the version line for --version'
        When call TestFindVersion
        The output should include 'find (mimixbox)'
        The status should be success
    End
End

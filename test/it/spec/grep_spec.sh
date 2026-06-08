Describe 'grep'
    Include findutils/grep_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'matches lines from stdin'
        When call TestGrepStdin
        The output should equal 'two'
        The status should be success
    End

    It 'matches lines from a file'
        When call TestGrepFile
        The output should equal 'banana'
        The status should be success
    End

    It 'counts matching lines with -c'
        When call TestGrepCount
        The output should equal '2'
        The status should be success
    End

    It 'exits 1 when nothing matches'
        When call TestGrepNoMatch
        The status should equal 1
    End
End

Describe 'script / scriptreplay'
    Include util-linux/script_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'records command output to a typescript'
        When call TestScriptRecords
        The output should equal '1'
        The status should be success
    End
    It 'replays a recorded typescript'
        When call TestScriptReplay
        The output should equal 'replayed'
        The status should be success
    End
End

Describe 'wall'
    Include util-linux/wall_test.sh

    It 'runs and exits successfully'
        When call TestWallRuns
        The output should equal 'rc=0'
        The status should be success
    End
End

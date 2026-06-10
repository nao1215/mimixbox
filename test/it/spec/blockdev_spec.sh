Describe 'blockdev'
    Include util-linux/blockdev_test.sh

    It 'describes itself with --help'
        When call TestBlockdevHelp
        The status should be success
        The output should include 'Usage: blockdev'
        The output should include 'DEVICE'
    End
    It 'fails when no query flag is given'
        When call TestBlockdevNoQuery
        The output should equal 'rc=1'
        The status should be success
    End
End

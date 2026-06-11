Describe 'mdev'
    Include util-linux/mdev_test.sh

    It 'requires scan mode'
        When call TestMdevNoScan
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestMdevHelp
        The status should be success
        The output should include 'Usage: mdev'
        The output should include 'device'
    End
End

Describe 'rdev'
    Include util-linux/rdev_test.sh

    It 'prints the root device with the / mountpoint'
        When call TestRdev
        The status should be success
        The output should include ' /'
    End
End

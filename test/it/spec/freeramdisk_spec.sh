Describe 'freeramdisk'
    Include util-linux/freeramdisk_test.sh

    It 'requires a device'
        When call TestFreeramdiskNoDev
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestFreeramdiskHelp
        The status should be success
        The output should include 'Usage: freeramdisk'
        The output should include 'ramdisk'
    End
End

Describe 'mountpoint'
    Include fileutils/mountpoint_test.sh
    It 'reports that / is a mountpoint'
        When call TestMountpointRoot
        The output should equal '/ is a mountpoint'
        The status should be success
    End
End

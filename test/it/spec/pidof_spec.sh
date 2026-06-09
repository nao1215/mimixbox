Describe 'pidof'
    Include shellutils/pidof_test.sh
    It 'finds the PID of a running process via MimixBox pidof'
        When call TestPidofInit
        The output should equal 'found'
        The status should be success
    End
    It 'resolves bare pidof to the MimixBox-installed symlink'
        When call TestPidofIsMimixBox
        The output should equal 'linked'
        The status should be success
    End
End

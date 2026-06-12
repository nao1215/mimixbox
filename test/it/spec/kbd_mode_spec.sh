Describe 'kbd_mode'
    Include console-tools/kbd_mode_test.sh

    It 'rejects conflicting mode options'
        When call TestKbdModeConflict
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestKbdModeHelp
        The status should be success
        The output should include 'Usage: kbd_mode'
        The output should include 'keyboard'
    End
End

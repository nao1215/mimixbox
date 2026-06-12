Describe 'setkeycodes'
    Include console-tools/setkeycodes_test.sh

    It 'requires arguments in pairs'
        When call TestSetkeycodesOdd
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects an invalid scancode'
        When call TestSetkeycodesBad
        The output should equal 'rc=1'
        The status should be success
    End
End

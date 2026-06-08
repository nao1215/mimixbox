Describe 'xxd dump'
    Include textutils/xxd_test.sh
    It 'prints a hex dump'
        When call TestXxdDump
        The output should equal '00000000: 6865 6c6c 6f0a                           hello.'
        The status should be success
    End
End

Describe 'xxd revert'
    Include textutils/xxd_test.sh
    It 'reverses a hex dump'
        When call TestXxdRevert
        The output should equal 'hello'
        The status should be success
    End
End

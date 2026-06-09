Describe 'hexdump / hd'
    Include util-linux/hexdump_test.sh

    It 'hd shows the canonical hex+ASCII layout'
        When call TestHd
        The status should be success
        The output should include '00000000  68 65 6c 6c 6f 20 77 6f  72 6c 64 0a'
        The output should include '|hello world.|'
    End
    It 'hexdump -C matches hd'
        When call TestHexdumpCanonical
        The status should be success
        The output should include '|hello world.|'
    End
    It 'hexdump default shows two-byte words'
        When call TestHexdumpDefault
        The status should be success
        The output should include '6568 6c6c 206f'
    End
End

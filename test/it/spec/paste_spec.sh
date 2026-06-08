Describe 'paste serial'
    Include textutils/paste_test.sh
    It 'joins lines with a delimiter'
        When call TestPasteSerial
        The output should equal 'a,b,c'
        The status should be success
    End
End

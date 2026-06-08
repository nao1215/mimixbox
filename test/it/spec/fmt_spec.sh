Describe 'fmt by width'
    Include textutils/fmt_test.sh
    result() { %text
        #|aa bb
        #|cc dd
    }
    It 'reflows text to the given width'
        When call TestFmtWidth
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'fold by width'
    Include textutils/fold_test.sh
    result() { %text
        #|abc
        #|def
        #|gh
    }
    It 'wraps lines to the given width'
        When call TestFoldWidth
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Remove one file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|2.txt
        #|3.txt
        #|inner
    }
    It 'remove one file.'
        When call TestRmOneFile
        The output should equal "$(result)"
    End
End

Describe 'Check status after removing one file'
    Include fileutils/rm_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'remove one file.'
        When call TestRmOneStatus
        The status should be success
    End
End

#Describe 'Remove three file using wildcard'
#    Include fileutils/rm_test.sh
#    BeforeEach 'Setup'
#    AfterEach 'Cleanup'
#
#    It 'remove three file.'
#        When call TestRmFileWithWildcard
#        The output should equal "inner"
#    End
#End
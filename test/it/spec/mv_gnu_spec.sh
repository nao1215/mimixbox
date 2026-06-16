Describe 'mv --target-directory moves operands into a directory'
    Include fileutils/mv_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|a.txt
        #|b.txt
    }

    It 'moves each source into the target directory'
        When call TestMvTargetDirectory
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'mv --update preserves a newer destination'
    Include fileutils/mv_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'does not overwrite a destination newer than the source'
        When call TestMvUpdateKeepsNewer
        The output should equal 'newer-dest'
        The status should be success
    End
End

Describe 'ln -s --relative creates a relative symbolic link'
    Include fileutils/ln_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'stores the target relative to the link location'
        When call TestLnRelative
        The output should equal '../a/target.txt'
        The status should be success
    End
End

Describe 'ln --target-directory links operands into a directory'
    Include fileutils/ln_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|a.txt
        #|b.txt
    }

    It 'creates a link per operand in the target directory'
        When call TestLnTargetDirectory
        The output should equal "$(result)"
        The status should be success
    End
End

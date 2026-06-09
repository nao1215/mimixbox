Describe 'more / less'
    Include console-tools/pager_test.sh

    It 'more streams stdin through when stdout is not a terminal'
        When call TestMorePassthrough
        The output should equal 'a
b
c'
        The status should be success
    End
    It 'less streams stdin through when stdout is not a terminal'
        When call TestLessPassthrough
        The output should equal 'x
y
z'
        The status should be success
    End
    It 'more streams a file through'
        When call TestMoreFile
        The output should equal 'one
two'
        The status should be success
    End
End

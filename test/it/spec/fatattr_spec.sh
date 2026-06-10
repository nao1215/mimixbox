Describe 'fatattr'
    Include util-linux/fatattr_test.sh

    It 'requires a file'
        When call TestFatattrNoFile
        The output should equal 'rc=1'
        The status should be success
    End
    It 'rejects an unknown attribute'
        When call TestFatattrBadAttr
        The output should equal 'rc=1'
        The status should be success
    End
End

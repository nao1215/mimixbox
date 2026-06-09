Describe 'cp symlink handling'
    Include fileutils/cp_symlink_test.sh

    It 'cp -P copies the symlink as a link'
        When call CpSymlinkP
        The status should be success
        The output should equal 'link'
    End

    It 'cp -L copies the link target as a regular file'
        When call CpSymlinkL
        The status should be success
        The output should equal 'regular'
    End

    It 'cp -d preserves a symlink inside a copied tree'
        When call CpSymlinkDInTree
        The status should be success
        The output should equal 'link'
    End
End

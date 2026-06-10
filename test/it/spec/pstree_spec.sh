Describe 'pstree'
    Include procps/pstree_test.sh

    It 'builds a tree containing PID 1'
        When call TestPstreeHasInit
        The status should be success
        The output should not equal '0'
    End
End

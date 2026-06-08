Describe 'uname'
    Include shellutils/uname_test.sh
    It 'prints the kernel name'
        When call TestUnameKernel
        The output should equal 'Linux'
        The status should be success
    End
End

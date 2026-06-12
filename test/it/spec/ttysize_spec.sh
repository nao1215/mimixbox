Describe 'ttysize'
    Include console-tools/ttysize_test.sh

    It 'prints width and height'
        When call TestTtysize
        The output should equal '80 24'
        The status should be success
    End
    It 'prints just the width with w'
        When call TestTtysizeWidth
        The output should equal '80'
        The status should be success
    End
End

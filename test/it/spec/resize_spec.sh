Describe 'resize prints usage'
    Include console-tools/resize_test.sh

    It 'shows the usage line for --help'
        When call TestResizeHelp
        The output should include 'Usage: resize'
        The status should be success
    End
End

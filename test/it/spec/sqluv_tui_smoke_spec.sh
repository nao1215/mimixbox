Describe 'sqluv TUI startup smoke'
    Include textutils/sqluv_test.sh
    It 'renders the minimal viewer and exits cleanly on quit'
        When call SqluvTUISmoke
        The status should be success
        The output should include 'minimal viewer'
        The output should include 'bye'
    End
End

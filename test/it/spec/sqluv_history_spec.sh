Describe 'sqluv configurable history file'
    Include textutils/sqluv_test.sh
    It 'writes query history to the path given by --history-file'
        When call SqluvHistory
        The status should be success
        The output should include 'select count(*) from nums'
    End
End

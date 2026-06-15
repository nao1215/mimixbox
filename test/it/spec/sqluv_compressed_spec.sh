Describe 'sqluv compressed local input'
    Include textutils/sqluv_test.sh
    It 'queries a gzip-compressed CSV fixture'
        When call SqluvCompressed
        The status should be success
        The output should include '2'
    End
End

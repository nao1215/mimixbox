Describe 'http-status-code'
    Include netutils/httpstatus_test.sh
    It 'explains a status code'
        When call TestHttpStatusSearch
        The output should include '404 Not Found'
        The status should be success
    End
End

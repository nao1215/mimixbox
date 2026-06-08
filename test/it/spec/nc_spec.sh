Describe 'nc loopback'
    Include netutils/nc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'transfers data over TCP'
        When call TestNcLoopback
        The output should equal 'from-client'
        The status should be success
    End
End

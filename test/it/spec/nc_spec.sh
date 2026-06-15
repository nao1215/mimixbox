Describe 'nc loopback'
    Include netutils/nc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'transfers data over TCP'
        # Sandboxed/locked-down hosts (seccomp, restricted netns, no
        # CAP_NET_RAW/BIND) forbid opening the loopback sockets this contract
        # needs and report "operation not permitted". That is an environment
        # capability gap, not an nc regression, so skip instead of failing.
        Skip if 'environment forbids opening the required sockets' NcSocketsForbidden
        When call TestNcLoopback
        The output should equal 'from-client'
        The status should be success
    End
End

Describe 'ipcs'
    Include util-linux/ipcs_test.sh

    It 'shows the IPC facility sections'
        When call TestIpcsAll
        The output should equal '1'
        The status should be success
    End
    It 'limits to shared memory with -m'
        When call TestIpcsShm
        The status should be success
        The output should include 'Shared Memory Segments'
        The output should not include 'Message Queues'
    End
End

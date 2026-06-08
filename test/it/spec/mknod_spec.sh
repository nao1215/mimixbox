Describe 'mknod creates special files'
    Include shellutils/mknod_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates a FIFO with type p'
        When call TestMknodFifo
        The output should equal 'fifo'
        The status should be success
    End

    It 'rejects an invalid device type'
        When call TestMknodBadType
        The status should be failure
        The error should include 'invalid device type'
    End
End

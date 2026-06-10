Describe 'fdisk'
    Include util-linux/fdisk_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/fdisk; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'lists an MBR Linux partition'
        When call TestFdiskList
        The output should equal '1'
        The status should be success
    End
    It 'rejects an image without an MBR signature'
        When call TestFdiskBad
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'fsck'
    Include util-linux/fsck_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/fsck; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'detects a Minix filesystem'
        When call TestFsckMinix
        The output should include 'minix'
        The status should be success
    End
    It 'fails on an unrecognized image'
        When call TestFsckUnknown
        The output should equal 'rc=1'
        The status should be success
    End
End

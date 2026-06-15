Describe 'fsck.minix'
    Include util-linux/fsck_minix_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/fsck_minix; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'validates a freshly made Minix filesystem'
        When call TestFsckMinixRoundTrip
        The output should include 'Minix v1'
        The status should be success
    End
    It 'rejects a non-Minix image'
        When call TestFsckMinixBad
        The output should equal 'rc=1'
        The status should be success
    End
End

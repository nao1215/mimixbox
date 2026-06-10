Describe 'mkfs.minix'
    Include util-linux/mkfs_minix_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/mkfs_minix; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'writes the Minix v1 magic'
        When call TestMkfsMinixMagic
        The output should equal '7f13'
        The status should be success
    End
    It 'refuses a too-small device'
        When call TestMkfsMinixTooSmall
        The output should equal 'rc=1'
        The status should be success
    End
End

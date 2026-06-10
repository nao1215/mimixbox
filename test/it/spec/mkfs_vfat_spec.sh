Describe 'mkfs.vfat / mkdosfs'
    Include util-linux/mkfs_vfat_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/mkfs_vfat; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'writes the FAT16 type label'
        When call TestMkfsVfatSig
        The output should equal 'FAT16'
        The status should be success
    End
    It 'mkdosfs refuses a too-small image'
        When call TestMkdosfsTooSmall
        The output should equal 'rc=1'
        The status should be success
    End
End

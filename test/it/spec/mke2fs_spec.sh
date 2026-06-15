Describe 'mke2fs / mkfs.ext2'
    Include util-linux/mke2fs_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/mke2fs; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'writes the ext2 magic'
        When call TestMke2fsMagic
        The output should equal '53ef'
        The status should be success
    End
    It 'mkfs.ext2 refuses an oversized image'
        When call TestMkfsExt2TooLarge
        The output should equal 'rc=1'
        The status should be success
    End
End

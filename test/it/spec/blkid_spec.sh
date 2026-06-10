Describe 'blkid'
    Include util-linux/blkid_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'identifies an ext filesystem'
        When call TestBlkidExt
        The output should include 'TYPE="ext2"'
        The status should be success
    End
    It 'identifies an xfs filesystem'
        When call TestBlkidXfs
        The output should include 'TYPE="xfs"'
        The status should be success
    End
    It 'exits 2 when nothing is identified'
        When call TestBlkidNone
        The output should equal 'rc=2'
        The status should be success
    End
End

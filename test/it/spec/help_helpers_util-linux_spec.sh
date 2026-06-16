# Per-command --help contract specs backed by dedicated util-linux helpers (issue #489).
Describe 'util-linux commands expose a dedicated --help helper'
    Include util-linux/fallocate_test.sh
    Include util-linux/fsck.minix_test.sh
    Include util-linux/linux32_test.sh
    Include util-linux/linux64_test.sh
    Include util-linux/mkdosfs_test.sh
    Include util-linux/mkfs.ext2_test.sh
    Include util-linux/mkfs.minix_test.sh
    Include util-linux/mkfs.reiser_test.sh
    Include util-linux/mkfs.vfat_test.sh
    Include util-linux/scriptreplay_test.sh
    Include util-linux/setsid_test.sh
    Include util-linux/sh_test.sh
    Include util-linux/swapoff_test.sh
    Include util-linux/swapon_test.sh

    It 'fallocate --help is structured'
        When call FallocateHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'fsck.minix --help is structured'
        When call FsckMinixHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'linux32 --help is structured'
        When call Linux32Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'linux64 --help is structured'
        When call Linux64Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'mkdosfs --help is structured'
        When call MkdosfsHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'mkfs.ext2 --help is structured'
        When call MkfsExt2Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'mkfs.minix --help is structured'
        When call MkfsMinixHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'mkfs.reiser --help is structured'
        When call MkfsReiserHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'mkfs.vfat --help is structured'
        When call MkfsVfatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'scriptreplay --help is structured'
        When call ScriptreplayHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'setsid --help is structured'
        When call SetsidHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'sh --help is structured'
        When call ShHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'swapoff --help is structured'
        When call SwapoffHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'swapon --help is structured'
        When call SwaponHelp
        The status should be success
        The output should include 'Usage:'
    End
End

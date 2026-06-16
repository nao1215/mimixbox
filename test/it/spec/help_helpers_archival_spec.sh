# Per-command --help contract specs backed by dedicated archival helpers (issue #489).
Describe 'archival commands expose a dedicated --help helper'
    Include archival/bzcat_test.sh
    Include archival/bzip2_test.sh
    Include archival/dpkg_test.sh
    Include archival/dpkg-deb_test.sh
    Include archival/lzcat_test.sh
    Include archival/lzma_test.sh
    Include archival/lzopcat_test.sh
    Include archival/pipe_progress_test.sh
    Include archival/rpm2cpio_test.sh
    Include archival/uncompress_test.sh
    Include archival/unlzma_test.sh
    Include archival/unlzop_test.sh
    Include archival/unxz_test.sh
    Include archival/unzip_test.sh
    Include archival/xz_test.sh
    Include archival/xzcat_test.sh
    Include archival/zcat_test.sh

    It 'bzcat --help is structured'
        When call BzcatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'bzip2 --help is structured'
        When call Bzip2Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'dpkg --help is structured'
        When call DpkgHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dpkg-deb --help is structured'
        When call DpkgDebHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lzcat --help is structured'
        When call LzcatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lzma --help is structured'
        When call LzmaHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'lzopcat --help is structured'
        When call LzopcatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'pipe_progress --help is structured'
        When call PipeProgressHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'rpm2cpio --help is structured'
        When call Rpm2cpioHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'uncompress --help is structured'
        When call UncompressHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'unlzma --help is structured'
        When call UnlzmaHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'unlzop --help is structured'
        When call UnlzopHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'unxz --help is structured'
        When call UnxzHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'unzip --help is structured'
        When call UnzipHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'xz --help is structured'
        When call XzHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'xzcat --help is structured'
        When call XzcatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'zcat --help is structured'
        When call ZcatHelp
        The status should be success
        The output should include 'Usage:'
    End
End

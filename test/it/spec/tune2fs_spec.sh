Describe 'tune2fs'
    Include util-linux/tune2fs_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/tune2fs; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'rejects a non-ext image'
        When call TestTune2fsNotExt
        The output should equal 'rc=1'
        The status should be success
    End
    It 'describes itself with --help'
        When call TestTune2fsHelp
        The status should be success
        The output should include 'Usage: tune2fs'
        The output should include 'filesystem'
    End
End

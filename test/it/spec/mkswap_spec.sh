Describe 'mkswap'
    Include util-linux/mkswap_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/mkswap; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'formats an image as swap'
        When call TestMkswapImage
        The output should equal '1'
        The status should be success
    End
    It 'writes the swap signature'
        When call TestMkswapSignature
        The output should not equal '0'
        The status should be success
    End
End

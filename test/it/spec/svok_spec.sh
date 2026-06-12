Describe 'svok'
    Include runit/svok_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/svok; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'succeeds for a supervised service'
        When call TestSvokSupervised
        The output should equal 'rc=0'
        The status should be success
    End
    It 'returns 100 for an unsupervised service'
        When call TestSvokNotSupervised
        The output should equal 'rc=100'
        The status should be success
    End
End

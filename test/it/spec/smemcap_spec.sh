Describe 'smemcap'
    Include procps/smemcap_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/smemcap; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'captures a tar containing meminfo'
        When call TestSmemcap
        The output should equal '1'
        The status should be success
    End
End

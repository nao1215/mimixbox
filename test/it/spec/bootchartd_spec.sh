Describe 'bootchartd'
    Include loginutils/bootchartd_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/bootchartd; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'records a proc_stat sample'
        When call TestBootchartdSample
        The output should not equal '0'
        The status should be success
    End
End

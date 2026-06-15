Describe 'svc'
    Include runit/svc_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/svc; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'writes the down control character'
        When call TestSvc
        The output should equal 'd'
        The status should be success
    End
    It 'requires a control command'
        When call TestSvcNoCmd
        The output should equal 'rc=1'
        The status should be success
    End
End

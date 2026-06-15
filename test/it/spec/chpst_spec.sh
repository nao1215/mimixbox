Describe 'chpst'
    Include runit/chpst_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/chpst; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'loads an environment directory'
        When call TestChpstEnvdir
        The output should equal 'world'
        The status should be success
    End
    It 'requires a program'
        When call TestChpstNoProg
        The output should equal 'rc=1'
        The status should be success
    End
End

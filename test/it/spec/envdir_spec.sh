Describe 'envdir'
    Include runit/envdir_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/envdir; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'sets a variable from a directory file'
        When call TestEnvdir
        The output should equal 'hello'
        The status should be success
    End
    It 'requires a directory and a program'
        When call TestEnvdirNoArgs
        The output should equal 'rc=1'
        The status should be success
    End
End

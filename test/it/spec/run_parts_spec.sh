Describe 'run-parts'
    Include loginutils/run_parts_test.sh

    setup() { TEST_DIR=${MIMIXBOX_IT_ROOT}/run_parts; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'runs executables in alphabetical order'
        When call TestRunPartsOrder
        The line 1 of output should equal 'A'
        The line 2 of output should equal 'B'
        The status should be success
    End
    It 'requires a directory'
        When call TestRunPartsNoDir
        The output should equal 'rc=1'
        The status should be success
    End
End

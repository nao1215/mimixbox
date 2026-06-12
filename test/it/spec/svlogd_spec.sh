Describe 'svlogd'
    Include runit/svlogd_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/svlogd; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'appends stdin to the current log'
        When call TestSvlogd
        The line 1 of output should equal 'hello'
        The line 2 of output should equal 'world'
        The status should be success
    End
    It 'requires a directory'
        When call TestSvlogdNoDir
        The output should equal 'rc=1'
        The status should be success
    End
End

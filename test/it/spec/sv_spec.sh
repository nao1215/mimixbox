Describe 'sv'
    Include runit/sv_test.sh

    setup() { TEST_DIR=/tmp/mimixbox/it/sv; mkdir -p "$TEST_DIR"; }
    cleanup() { rm -rf "$TEST_DIR"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'writes the up control character'
        When call TestSvUp
        The output should equal 'u'
        The status should be success
    End
    It 'reports a running service'
        When call TestSvStatus
        The output should include 'run'
        The status should be success
    End
End

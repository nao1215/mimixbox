Describe 'readlink'
    Include fileutils/readlink_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'prints the symlink target'
        When call TestReadlink
        The output should equal '/tmp/mimixbox/it/rl_target'
        The status should be success
    End
End

Describe 'sync'
    Include shellutils/sync_test.sh
    It 'flushes filesystem buffers'
        When call TestSync
        The output should equal 'synced'
        The status should be success
    End
End

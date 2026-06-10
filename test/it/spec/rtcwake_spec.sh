Describe 'rtcwake'
    Include util-linux/rtcwake_test.sh

    It 'rejects a suspend mode'
        When call TestRtcwakeSuspendRejected
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a wake time'
        When call TestRtcwakeNoTime
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'mimixbox top-level CLI'
    Include shellutils/mimixbox_test.sh

    It 'prints usage to stdout with --help and exits success'
        When call MbHelp
        The status should be success
        The output should include 'Usage: mimixbox'
        The output should include 'Examples:'
    End

    It 'lists the applets with --list'
        When call MbList
        The status should be success
        The output should include 'cat'
        The output should include 'pidof'
    End

    It 'rejects an unknown option on stderr without polluting stdout'
        When call MbUnknownOption
        The status should be failure
        The output should equal ''
        The error should include 'is not a mimixbox command or option'
    End

    It 'installs and removes applet symlinks in a temp directory'
        When call MbInstallRemoveSmoke
        The status should be success
        The output should equal 'ok'
    End
End

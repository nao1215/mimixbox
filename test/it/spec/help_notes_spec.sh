Describe 'applets with a Notes section in --help'
    # GitHub issues #709-#720: every listed applet must route --help through the
    # structured help renderer and include a "Notes:" section, exiting 0.

    Parameters
        acpid
        brctl
        crond
        ifenslave
        mkfs.reiser
        nbd-client
        ssl_server
        tunctl
        vconfig
        zcip
    End

    It "$1 --help documents a Notes section"
        When run "$1" --help
        The status should be success
        The output should include 'Notes:'
        The line 1 of output should start with "Usage: $1"
    End

    # `[` and `[[` collide with shell builtins, so run them through env to force
    # the applet on PATH (GitHub issues #709, #710).
    It '[ --help documents a Notes section'
        When run env '[' --help
        The status should be success
        The output should include 'Notes:'
        The line 1 of output should start with 'Usage: ['
    End

    It '[[ --help documents a Notes section'
        When run env '[[' --help
        The status should be success
        The output should include 'Notes:'
        The line 1 of output should start with 'Usage: [['
    End
End

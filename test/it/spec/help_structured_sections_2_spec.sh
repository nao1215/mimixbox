Describe 'more applets expose structured --help sections'
    # GitHub issues: "add structured help sections for X" (second alphabet half).
    # Each applet must route --help through the structured renderer, exit 0, and
    # document a purpose paragraph, an Examples section, and an Exit status section.

    Parameters
        add-shell
        ar
        arch
        awk
        banner
        base32
        base64
        basename
        bunzip2
        cal
        cat
        chgrp
        chmod
        chown
        cksum
        clear
        cmatrix
        cmp
        comm
        compress
        cowsay
        cowthink
        cpio
        cut
        date
        dd
        df
        diff
        dirname
        dos2unix
        du
        egrep
        env
        expand
        expr
        fakemovie
        fgrep
        fmt
        fold
        fortune
        free
        ghrdc
        grep
        groups
        gunzip
        gzip
        halt
        head
        hostid
        hostname
        http-status-code
        id
        install
        ischroot
        killall
        lifegame
        link
    End

    It "$1 --help exposes structured sections"
        When run "$1" --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with "Usage: $1"
        The output should include "  $1 "
    End

    # echo/false/kill are shell builtins; run them through env to reach the applet.
    It 'echo --help exposes structured sections'
        When run env echo --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: echo'
        The output should include '  echo '
    End

    It 'false --help exposes structured sections'
        When run env false --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: false'
        The output should include '  false '
    End

    It 'kill --help exposes structured sections'
        When run env kill --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: kill'
        The output should include '  kill '
    End

End

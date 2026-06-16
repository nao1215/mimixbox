Describe 'applets expose structured --help sections'
    # GitHub issues: "add structured help sections for X". Each applet must
    # route --help through the structured renderer, exit 0, and document a
    # purpose paragraph, an Examples section, and an Exit status section.

    Parameters
        ln
        log-collect
        logname
        md5sum
        mkdir
        mkfifo
        mknod
        mktemp
        mountpoint
        mv
        nc
        netcat
        nl
        nohup
        nproc
        nyancat
        od
        paste
        patch
        path
        pidof
        ping
        posixer
        poweroff
        printenv
        pwcrack
        pwgen
        pwscore
        readlink
        realpath
        reboot
        remove-shell
        reset
        resize
        rev
        rm
        rmdir
        rpm
        rpm2cpio
        sddf
        sed
        seq
        serial
        sha1sum
        sha256sum
        sha384sum
        sha3sum
        sha512sum
        shred
        shuf
        sl
        sleep
        sort
        speaker
        split
        stat
        strings
        sync
        tac
        tar
        tee
        timeout
        touch
        tr
        truncate
        tty
        uname
        uncompress
        unexpand
        uniq
        unix2dos
        unlink
        unshadow
        unzip
        uuidgen
        valid-shell
        watch
        wc
        which
        who
        whoami
        whris
        xargs
        xxd
        yes
        zip
        zip-pwcrack
    End

    It "$1 --help exposes structured sections"
        When run "$1" --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with "Usage: $1"
        # At least one example command line starts with the command name.
        The output should include "  $1 "
    End

    # true/test/printf/pwd are shell builtins, so run them through env to
    # force the applet on PATH.
    It 'true --help exposes structured sections'
        When run env true --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: true'
        The output should include '  true '
    End

    It 'test --help exposes structured sections'
        When run env test --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: test'
        The output should include '  test '
    End

    It 'printf --help exposes structured sections'
        When run env printf --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: printf'
        The output should include '  printf '
    End

    It 'pwd --help exposes structured sections'
        When run env pwd --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with 'Usage: pwd'
        The output should include '  pwd '
    End

End

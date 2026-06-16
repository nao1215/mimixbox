Describe 'applets document their exit status in --help'
    # GitHub issues #661-#697: every listed applet must route --help through the
    # structured help renderer and include an "Exit status:" section, exiting 0.

    Parameters
        ash
        bash
        bc
        busybox
        cttyhack
        dc
        ed
        hd
        hexdump
        hush
        iostat
        ipcs
        last
        less
        lsblk
        lspci
        lsusb
        mbsh
        minips
        more
        mpstat
        nmeter
        pipe_progress
        powertop
        ps
        pstree
        sh
        smemcap
        top
        uptime
        users
        uudecode
        uuencode
        vi
        vmstat
        w
        wall
    End

    It "$1 --help documents its exit status"
        When run "$1" --help
        The status should be success
        The output should include 'Exit status:'
        The line 1 of output should start with "Usage: $1"
    End
End

Describe 'compression applets self-describing --help'
    # GitHub issues #649-#660 (Examples) and #698-#708 (purpose paragraph):
    # every compression/decompression applet must route --help through the
    # structured help renderer, exiting 0 with an Examples and Exit status
    # section and an example line that starts with the command name.

    Parameters
        xz
        unxz
        xzcat
        lzma
        unlzma
        lzcat
        lzop
        unlzop
        lzopcat
        zcat
        bzcat
        unit
    End

    It "$1 --help has examples and an exit-status section"
        When run "$1" --help
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
        The line 1 of output should start with "Usage: $1"
        # A non-empty purpose paragraph sits on line 3 (line 2 is blank).
        The line 3 of output should not equal ''
        # At least one example command line starts with the command name.
        The output should include "  $1 "
    End
End

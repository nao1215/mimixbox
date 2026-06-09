Describe 'hermetic harness'
    Include shellutils/hermetic_test.sh

    # These applets all have host equivalents (coreutils), so if the suite were
    # not hermetic they would resolve to the host binary and these cases would
    # fail. cat/head/base64 are external programs, not shell builtins, so
    # `command -v` returns a real path that we can follow to the MimixBox binary.
    Parameters
        cat
        head
        base64
    End

    It "resolves $1 to the MimixBox binary, not the host command"
        When call ResolvesToMimixBox "$1"
        The output should equal 'mimixbox'
        The status should be success
    End
End

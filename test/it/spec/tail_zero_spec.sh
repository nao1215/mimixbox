Describe 'tail --zero-terminated'
    # The fixture holds two NUL-delimited records, each containing an embedded
    # newline: record 1 is "a\nb" and record 2 is "c\nd". With -z, tail must
    # split on NUL (not newline) and keep the embedded newlines verbatim, and
    # emit NUL as the trailing record separator.
    Describe 'last NUL record with -z'
        # tail -z -n 1 -> "c\nd\0"; rendering NUL as '|' gives "c\nd|".
        result() { %text
            #|c
            #|d|
        }

        It 'prints the last NUL-delimited record, preserving the embedded newline'
            When call sh -c "printf 'a\nb\0c\nd\0' | tail -z -n 1 | tr '\0' '|'"
            The output should equal "$(result)"
            The status should be success
        End
    End

    Describe 'two NUL records with --zero-terminated'
        # tail --zero-terminated -n 2 -> "a\nb\0c\nd\0" -> "a\nb|c\nd|".
        result() { %text
            #|a
            #|b|c
            #|d|
        }

        It 'prints two NUL-delimited records with embedded newlines preserved'
            When call sh -c "printf 'a\nb\0c\nd\0' | tail --zero-terminated -n 2 | tr '\0' '|'"
            The output should equal "$(result)"
            The status should be success
        End
    End
End

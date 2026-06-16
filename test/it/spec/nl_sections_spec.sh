Describe 'nl numbers each header/body/footer section with its own style'
    Include textutils/nl_test.sh
    BeforeEach 'SetupNlSections'
    AfterEach 'Cleanup'

    result() {
        printf '%s\n' \
          "     1	H1" \
          "" \
          "     1	HDR" \
          "" \
          "     1	B1" \
          "" \
          "     1	F1"
    }
    It 'numbers every line in every section with -h a -b a -f a'
        When call TestNlSectionsAll
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'nl honors a different style per section'
    Include textutils/nl_test.sh
    BeforeEach 'SetupNlSections'
    AfterEach 'Cleanup'

    result() {
        printf '%s\n' \
          "     1	H1" \
          "" \
          "     1	HDR" \
          "" \
          "     1	B1" \
          "" \
          "       F1"
    }
    It 'numbers header (a) and body (t) but not footer (n)'
        When call TestNlSectionsMixed
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'nl counts a group of blank lines as one with -l'
    Include textutils/nl_test.sh
    BeforeEach 'SetupNlSections'
    AfterEach 'Cleanup'

    result() {
        printf '%s\n' \
          "     1	a" \
          "       " \
          "     2	" \
          "       " \
          "     3	" \
          "     4	b"
    }
    It 'numbers every second blank line with -l 2'
        When call TestNlJoinBlankLines
        The output should equal "$(result)"
        The status should be success
    End
End

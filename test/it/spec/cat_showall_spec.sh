Describe 'cat -A and --show-all are aliases'
    Include textutils/cat_showall_test.sh
    BeforeEach 'SetupShowAll'
    AfterEach 'CleanUpShowAll'

    It 'produces byte-identical output for -A and --show-all'
        When call TestCatShowAllAlias
        The output should equal "identical"
        The status should be success
    End
End

Describe 'cat -v and --show-nonprinting are aliases'
    Include textutils/cat_showall_test.sh
    BeforeEach 'SetupShowAll'
    AfterEach 'CleanUpShowAll'

    It 'produces byte-identical output for -v and --show-nonprinting'
        When call TestCatShowNonprintingAlias
        The output should equal "identical"
        The status should be success
    End
End

Describe 'cat --show-all rendering'
    Include textutils/cat_showall_test.sh
    BeforeEach 'SetupShowAll'
    AfterEach 'CleanUpShowAll'

    result() { %text
        #|a^Ib^A$
        #|$
        #|^?M-^@$
    }

    It 'shows tabs as ^I, non-printing bytes, and $ line ends'
        When call TestCatShowAllRendered
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'cat --show-nonprinting rendering'
    Include textutils/cat_showall_test.sh
    BeforeEach 'SetupShowAll'
    AfterEach 'CleanUpShowAll'

    # TAB is left untouched by -v; only control/DEL/high bytes are rendered.
    result() { printf 'a\tb^A\n\n^?M-^@'; }

    It 'leaves TAB alone and renders ^X, ^?, and M- notation'
        When call TestCatShowNonprintingRendered
        The output should equal "$(result)"
        The status should be success
    End
End

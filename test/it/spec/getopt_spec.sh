Describe 'getopt'
    Include util-linux/hexdump_test.sh

    It 'normalizes short and long options with quoted args'
        When call TestGetopt
        The status should be success
        The output should include "-b 'val'"
        The output should include "--alpha"
        The output should include "-- 'pos'"
    End
    It 'produces output a script can eval'
        When call TestGetoptEval
        The status should be success
        The output should equal '-n|hello|--|world|'
    End
End

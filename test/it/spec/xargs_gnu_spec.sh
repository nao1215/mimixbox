Describe 'xargs GNU flags'
    Include findutils/xargs_gnu_test.sh

    It 'runs once per input line with -L 1'
        When call TestXargsMaxLinesOne
        The output should equal '3'
        The status should be success
    End

    It 'groups two input lines per invocation with -L 2'
        When call TestXargsMaxLinesTwo
        The output should equal '2'
        The status should be success
    End

    It 'splits a long input into multiple invocations with -s'
        When call TestXargsMaxCharsSplits
        The output should equal '2'
        The status should be success
    End

    It 'keeps every item across the -s split invocations'
        When call TestXargsMaxCharsKeepsAllItems
        The output should equal '8'
        The status should be success
    End

    It 'runs all batches concurrently with -P 4'
        When call TestXargsMaxProcsAllRun
        The output should equal 'a b c d'
        The status should be success
    End

    It 'runs all batches with -P 0 (as many as possible)'
        When call TestXargsMaxProcsZero
        The output should equal 'a b c d'
        The status should be success
    End
End

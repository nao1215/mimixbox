Describe 'sort -V version sort'
    Include shellutils/sort_test.sh

    result() { %text
        #|1.1
        #|1.2
        #|1.10
    }

    It 'orders version numbers by value'
        When call TestSortVersion
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort -g general numeric sort'
    Include shellutils/sort_test.sh

    result() { %text
        #|0.5
        #|2.5
        #|100
        #|1e3
    }

    It 'orders floating-point values including exponents'
        When call TestSortGeneralNumeric
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort -h human numeric sort'
    Include shellutils/sort_test.sh

    result() { %text
        #|2K
        #|1M
        #|1G
    }

    It 'orders human-readable sizes by magnitude'
        When call TestSortHumanNumeric
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort -s stable sort'
    Include shellutils/sort_test.sh

    result() { %text
        #|5 zebra
        #|5 apple
        #|5 mango
    }

    It 'keeps input order for equal keys'
        When call TestSortStable
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort -z zero-terminated'
    Include shellutils/sort_test.sh

    It 'reads and writes NUL-delimited records'
        When call TestSortZeroTerminated
        The output should equal 'apple|banana|cherry|'
        The status should be success
    End
End

Describe 'sort -m merge'
    Include shellutils/sort_test.sh

    result() { %text
        #|apple
        #|banana
        #|cherry
    }

    It 'merges already-sorted input'
        When call TestSortMerge
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort --parallel accepted'
    Include shellutils/sort_test.sh

    result() { %text
        #|a
        #|b
    }

    It 'accepts --parallel without error'
        When call TestSortParallel
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'sort --temporary-directory accepted'
    Include shellutils/sort_test.sh

    result() { %text
        #|a
        #|b
    }

    It 'accepts --temporary-directory without error'
        When call TestSortTemporaryDirectory
        The output should equal "$(result)"
        The status should be success
    End
End

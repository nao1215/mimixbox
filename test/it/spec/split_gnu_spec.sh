Describe 'split GNU flags'
    Include textutils/split_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'uses numeric suffixes with -d'
        When call TestSplitNumeric
        The output should equal 'num-00 num-01 num-02'
        The status should be success
    End

    It 'writes the expected content to the first numeric piece'
        When call TestSplitNumericContent
        The output should equal "$(printf '1\n2')"
        The status should be success
    End

    It 'appends an additional suffix to each name'
        When call TestSplitAdditionalSuffix
        The output should equal 'add-aa.txt add-ab.txt'
        The status should be success
    End

    It 'honors a custom suffix length with -a'
        When call TestSplitSuffixLength
        The output should equal 'len-aaa len-aab'
        The status should be success
    End
End

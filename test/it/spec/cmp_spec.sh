Describe 'cmp on identical files'
    Include shellutils/cmp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints nothing and succeeds'
        When call TestCmpEqual
        The output should equal 'rc=0'
        The status should be success
    End
End

Describe 'cmp on differing files'
    Include shellutils/cmp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports the first differing byte and line'
        When call TestCmpDiffer
        The output should equal "${MIMIXBOX_IT_ROOT}/cmp/a.txt ${MIMIXBOX_IT_ROOT}/cmp/diff.txt differ: byte 3, line 1"
        The status should be failure
    End
End

Describe 'cmp -s on differing files'
    Include shellutils/cmp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints nothing but exits non-zero'
        When call TestCmpSilent
        The output should equal 'rc=1'
        The status should be success
    End
End

Describe 'cp preserves permissions'
    Include shellutils/cp_perm_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'keeps the source file mode (execute bit)'
        When call TestCpPreservesExecBit
        The output should equal '755'
        The status should be success
    End
    It 'keeps a private directory mode'
        When call TestCpPreservesDirMode
        The output should equal '700'
        The status should be success
    End
    It 'overwrites a read-only destination with -f'
        When call TestCpForceOverwritesReadonly
        The output should equal 'new'
        The status should be success
    End
End

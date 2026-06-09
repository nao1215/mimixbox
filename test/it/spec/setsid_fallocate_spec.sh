Describe 'setsid / fallocate'
    Include util-linux/setsid_fallocate_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'setsid runs a program in a new session'
        When call TestSetsid
        The output should equal 'session ok'
        The status should be success
    End
    It 'fallocate sizes a file to the requested length'
        When call TestFallocate
        The status should be success
        The output should include '4096'
    End
    It 'fallocate without -l fails'
        When call TestFallocateNoLength
        The output should equal '1'
        The status should be success
    End
End

Describe 'ls'
    Include fileutils/ls_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'lists entries sorted, hiding dotfiles'
        When call TestLsDefault
        The output should equal 'a.txt
b.txt
sub'
        The status should be success
    End
    It 'includes dotfiles with -a'
        When call TestLsAll
        The output should include '.hidden'
        The status should be success
    End
    It 'marks directories with -F'
        When call TestLsClassify
        The output should include 'sub/'
        The status should be success
    End
End

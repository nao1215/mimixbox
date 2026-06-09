Describe 'coreutils slice'
    Include coreutils/coreutils_test.sh

    It 'factor prints prime factors'
        When call TestFactor
        The output should equal '360: 2 2 2 3 3 5'
        The status should be success
    End
    It 'tsort topologically sorts'
        When call TestTsort
        The output should equal 'a
b
c'
        The status should be success
    End
    It 'egrep uses extended regular expressions'
        When call TestEgrep
        The output should equal 'bar
baz'
        The status should be success
    End
    It 'fgrep matches fixed strings literally'
        When call TestFgrep
        The output should equal 'a.b'
        The status should be success
    End
End

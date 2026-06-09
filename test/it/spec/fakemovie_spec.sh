Describe 'fakemovie CLI contract'
    Include jokeutils/fakemovie_test.sh
    It 'prints usage with --help and exits 0'
        When call FakemovieHelp
        The status should be success
        The output should include 'Usage: fakemovie'
    End
    It 'fails with a message when given no operand'
        When call FakemovieNoArg
        The status should be failure
        The error should include 'fakemovie'
    End
End

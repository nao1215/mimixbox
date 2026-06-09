Describe 'cowsay CLI contract'
    Include jokeutils/cowsay_test.sh
    It 'prints usage with --help and exits 0'
        When call CowsayHelp
        The status should be success
        The output should include 'Usage: cowsay'
    End
    It 'renders the message in the speech bubble'
        When call CowsaySpeak
        The status should be success
        The output should include 'hello'
    End
End

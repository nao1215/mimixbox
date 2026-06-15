Describe 'leadtime CLI contract'
    Include shellutils/leadtime_test.sh
    It 'prints usage with --help and exits 0'
        When call LeadtimeHelp
        The status should be success
        The output should include 'Usage: leadtime'
        The output should include 'LT_GITHUB_ACCESS_TOKEN'
    End
    It 'fails when no subcommand is given'
        When call LeadtimeNoSub
        The status should be failure
        The error should include 'stat'
    End
    It 'fails on an unknown subcommand'
        When call LeadtimeUnknownSub
        The status should be failure
        The error should include 'unknown subcommand'
    End
    It 'fails with a deterministic error when no token is set'
        When call LeadtimeMissingToken
        The status should be failure
        The error should include 'no GitHub token'
    End
    It 'fails when --owner/--repo are missing'
        When call LeadtimeMissingOwnerRepo
        The status should be failure
        The error should include '--owner and --repo are required'
    End
    It 'rejects --json with --markdown'
        When call LeadtimeJSONMarkdownConflict
        The status should be failure
        The error should include 'mutually exclusive'
    End
End

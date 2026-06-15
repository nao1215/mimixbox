Describe 'sqluv headless contract'
    Include textutils/sqluv_test.sh
    It 'prints usage with --help and exits 0'
        When call SqluvHelp
        The status should be success
        The output should include 'Usage: sqluv'
        The output should include 'headless'
    End
    It 'prints the version and exits 0'
        When call SqluvVersion
        The status should be success
        The output should include 'sqluv (mimixbox)'
    End
    It 'fails with a message when given no operand'
        When call SqluvNoArg
        The status should be failure
        The error should include 'sqluv'
    End
    It 'queries a CSV fixture in headless mode'
        When call SqluvCSV
        The status should be success
        The output should include 'alice'
        The output should include 'bob'
    End
    It 'queries a SQLite-style table as JSON'
        When call SqluvSQLite
        The status should be success
        The output should include 'go-in-action'
    End
    It 'fails deterministically on an unsupported S3 source'
        When call SqluvUnsupported
        The status should be failure
        The error should include 'S3 sources are not migrated'
    End
End

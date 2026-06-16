Describe 'mimixbox top-level list/suggestion CLI'
    Include shellutils/mimixbox_test.sh

    # Issue #782: --list --json emits a JSON array including cat and ls.
    Describe '--list --json'
        It 'emits a JSON array containing cat and ls on stdout'
            When call MbListJSON
            The status should be success
            The output should include '"name": "cat"'
            The output should include '"name": "ls"'
            The output should include '"subsystem":'
            The output should include '"stability":'
            The output should include '['
            The output should include ']'
        End
    End

    # Issue #783: --list filtering by prefix and by subsystem.
    Describe '--list filtering'
        It 'includes cat and excludes ls with --filter=cat'
            When call MbListFilter
            The status should be success
            The output should include 'cat'
            The line 1 of output should include 'cat'
        End

        It 'includes cat and excludes ls with --subsystem=textutils'
            When call MbListSubsystem
            The status should be success
            The output should include 'cat'
            # ls lives in fileutils, so the textutils listing must not name it.
            The output should not include ' ls -'
        End
    End

    # Issue #781: an unknown command suggests the nearest applet, error-first.
    Describe 'unknown command suggestions'
        It "suggests ls for 'lss' before the full applet wall"
            When call MbUnknownCommand
            The status should be failure
            The output should equal ''
            The error should include "'lss' is not a mimixbox command."
            The error should include 'Did you mean:'
            The error should include 'ls'
        End
    End
End

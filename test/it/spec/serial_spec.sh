Describe 'serial renames files in a directory'
    Include fileutils/serial_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|0_apple.txt
        #|1_banana.txt
        #|2_cherry.txt
    }

    It 'adds a serial-number prefix to each file'
        When call TestSerialRename
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'serial with --dry-run'
    Include fileutils/serial_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
        #|apple.txt
        #|banana.txt
        #|cherry.txt
    }

    It 'does not rename anything'
        When call TestSerialDryRun
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'serial with no operand'
    Include fileutils/serial_test.sh

    It 'reports an error'
        When call TestSerialNoOperand
        The error should equal 'serial: missing operand'
        The status should be failure
    End
End

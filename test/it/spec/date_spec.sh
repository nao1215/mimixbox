Describe 'date with an epoch and format'
    Include shellutils/date_test.sh

    It 'formats the date portion'
        When call TestDateEpochDate
        The output should equal '1970-01-01'
        The status should be success
    End

    It 'formats the time portion'
        When call TestDateEpochTime
        The output should equal '00:00:00'
        The status should be success
    End
End

Describe 'date with a literal percent'
    Include shellutils/date_test.sh

    It 'prints a percent sign'
        When call TestDatePercent
        The output should equal '%'
        The status should be success
    End
End

Describe 'date year'
    Include shellutils/date_test.sh

    It 'prints a four-digit year'
        When call TestDateYearDigits
        The output should equal 'ok'
        The status should be success
    End
End

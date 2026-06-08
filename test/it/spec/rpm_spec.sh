Describe 'rpm and rpm2cpio'
    Include archival/rpm_test.sh

    It 'queries the package identity with rpm -qp'
        When call TestRpmQuery
        The output should equal 'hello-2.10-1.fc40.x86_64'
        The status should be success
    End
    It 'lists package files with rpm -qpl'
        When call TestRpmList
        The line 1 of output should equal '/usr/bin/hello'
        The line 2 of output should equal '/etc/hello.conf'
        The status should be success
    End
    It 'extracts the payload with rpm2cpio'
        When call TestRpm2cpio
        The output should equal 'RPM-PAYLOAD'
        The status should be success
    End
End

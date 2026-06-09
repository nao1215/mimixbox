Describe 'setarch / linux32 / linux64'
    Include util-linux/setarch_test.sh

    It 'linux32 makes uname report a 32-bit machine'
        When call TestLinux32Uname
        The output should equal 'i686'
        The status should be success
    End
    It 'linux64 reports the native machine'
        When call TestLinux64Uname
        The output should equal 'x86_64'
        The status should be success
    End
    It 'linux32 passes the command output through'
        When call TestLinux32Passthrough
        The output should equal 'passed'
        The status should be success
    End
    It 'setarch selects the personality from ARCH'
        When call TestSetarchArch
        The output should equal 'i686'
        The status should be success
    End
End

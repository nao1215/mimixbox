Describe 'iostat'
    Include procps/iostat_test.sh

    It 'prints the avg-cpu header'
        When call TestIostatCpuHeader
        The output should include 'avg-cpu'
        The output should include '%idle'
        The status should be success
    End
    It 'prints the device table header'
        When call TestIostatDeviceHeader
        The output should equal '1'
        The status should be success
    End
End

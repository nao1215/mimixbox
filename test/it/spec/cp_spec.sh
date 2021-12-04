Describe 'Copy one file'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'copy one file.'
        When call TestCopyOneFile
        The output should equal '/tmp/mimixbox/it/cp/cp.txt'
    End
End

Describe 'Check status after copying one file'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestCopyOneFile
        The output should equal '/tmp/mimixbox/it/cp/cp.txt'
        The status should be success
    End
End

Describe 'Copy directory'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|cp
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
    }

    It 'copy directory recursively'
        When call TestCopyOndDirWithRecursiveOption
        The output should equal "$(result)"
    End
End

Describe 'Check status after copying directory'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'


    It 'copy directory'
        When call TestCopyOndDirWithRecursiveOptionStatus
        The status should be success
    End
End

Describe 'The reason why the copy failed is src and dest are the same'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
        #|inner
    }

    It 'can not copy directory'
        When call TestCopySrcAddDistAreSame
        The output should equal "$(result)"
        The error should equal "cp: /tmp/mimixbox/it/cp and /tmp/mimixbox/it/cp is same."
    End
End

Describe 'Check status after copying fail. Src and dest are the same'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'can not copy directory'
        When call TestCopySrcAddDistAreSameStatus
        The error should equal "cp: /tmp/mimixbox/it/cp and /tmp/mimixbox/it/cp is same."
        The status should be failure
    End
End

Describe 'Copy three file at same time'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1.txt
        #|2.txt
        #|3.txt
    }

    It 'make three file'
        When call TestCopyThreeFileAtSameTime
        The output should equal "$(result)"
    End
End

Describe 'Check status after copying three file'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestCopyThreeFileAtSameTimeStatus
        The status should be success
    End
End

Describe 'Copy directory without recursive option'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'can not copy directory'
        When call TestCopyDirctoryWithoutRecursiveOption
        The output should equal ""
        The error should equal "cp: --recursive is not specified: omitting directory: /tmp/mimixbox/it/cp"
    End
End

Describe 'Check status after copy directory without recursive option'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status failure'
        When call TestCopyDirctoryWithoutRecursiveOptionStatus
        The error should equal "cp: --recursive is not specified: omitting directory: /tmp/mimixbox/it/cp"
        The status should be failure
    End
End

Describe 'Copy directory to Root'
    Include fileutils/cp_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'can not copy directory because do not have the authority'
        When call TestCopyDirectoryAtRoot
        The error should equal "cp: mkdir /cp: permission denied"
        The status should be failure
    End
End

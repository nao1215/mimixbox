Describe 'mktemp creates a file'
    Include debianutils/mktemp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates a regular file under the temp dir'
        When call TestMktempFile
        The output should equal 'created'
        The status should be success
    End
End

Describe 'mktemp -d creates a directory'
    Include debianutils/mktemp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates a directory'
        When call TestMktempDir
        The output should equal 'created'
        The status should be success
    End
End

Describe 'mktemp -u does not create the file'
    Include debianutils/mktemp_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'only prints a name'
        When call TestMktempDryRun
        The output should equal 'not created'
        The status should be success
    End
End

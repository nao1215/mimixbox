Describe 'env GNU chdir/split-string/ignore-signal flags (issue #756)'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/env_gnu"
        mkdir -p "$WORK/sub"
    }
    cleanup() { rm -rf "${MIMIXBOX_IT_ROOT}/env_gnu"; }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    Describe '--chdir runs the command in DIR'
        It 'reports the chdir target via pwd (long flag with =)'
            When run env --chdir=/tmp pwd
            The status should be success
            The output should equal '/tmp'
        End

        It 'reports the chdir target via pwd (-C short flag)'
            When run env -C "$WORK/sub" pwd
            The status should be success
            The output should equal "$WORK/sub"
        End

        It 'fails when the directory does not exist'
            When run env --chdir="$WORK/nope" pwd
            The status should equal 125
            The stderr should include 'cannot change directory'
        End
    End

    Describe '--split-string expands one string into multiple argv'
        It 'splits the command and its arguments (-S)'
            When run env -S 'printf %s-%s a b'
            The status should be success
            The output should equal 'a-b'
        End

        It 'splits with the long flag and whitespace runs'
            When run env --split-string='printf   %s   hi'
            The status should be success
            The output should equal 'hi'
        End
    End

    Describe '--ignore-signal validates signal names'
        It 'accepts known names and still runs the command'
            When run env --ignore-signal=INT,TERM printf '%s' ok
            The status should be success
            The output should equal 'ok'
        End

        It 'rejects an unknown signal name'
            When run env --ignore-signal=BOGUS printf x
            The status should equal 125
            The stderr should include 'invalid signal'
        End
    End
End

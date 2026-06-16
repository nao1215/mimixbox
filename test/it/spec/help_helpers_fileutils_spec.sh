# Per-command --help contract specs backed by dedicated fileutils helpers (issue #489).
Describe 'fileutils commands expose a dedicated --help helper'
    Include fileutils/chgrp_test.sh
    Include fileutils/chown_test.sh

    It 'chgrp --help is structured'
        When call ChgrpHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'chown --help is structured'
        When call ChownHelp
        The status should be success
        The output should include 'Usage:'
    End
End

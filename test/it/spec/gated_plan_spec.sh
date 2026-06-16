# shellcheck shell=sh
# Issue #788: deterministic planned-action E2E for privileged, capability-gated
# applets.
#
# The "never ship a silent no-op" rule means each privileged applet first
# validates its arguments and serializes the requested action (a plan), then
# fails with a documented capability/gate error and a non-zero exit WITHOUT
# performing the action. This spec picks one representative command per gated
# family (netctl, selinux, modutils) and asserts both the normalized plan/
# validation text and the gated error message, copied verbatim from each
# applet's source so the assertions are deterministic. One case per family.
Describe 'capability-gated applets validate, plan, then refuse'

  Describe 'netctl: brctl addbr br0'
    # internal/applets/netutils/netctl: a plan line on stdout, then a
    # capability-gated backend error; exit 1 and no bridge created.
    It 'prints the plan then fails with a capability-gated backend error'
      When run brctl addbr br0
      The status should be failure
      The output should include 'brctl: planned action: brctl addbr br0'
      The stderr should include 'planned action [brctl addbr br0] requires privileged kernel network configuration not available in this environment (capability-gated backend)'
    End
  End

  Describe 'selinux: setenforce Permissive'
    # internal/applets/securityutils/selinux: the mutating SELinux applets
    # refuse deterministically (CAP_MAC_ADMIN gate) instead of changing state.
    It 'refuses to mutate SELinux state and exits non-zero'
      When run setenforce Permissive
      The status should be failure
      The stderr should include 'setenforce: refusing to mutate SELinux state: requires CAP_MAC_ADMIN and a loaded policy; this operation is intentionally not implemented in the hermetic build'
    End
  End

  Describe 'modutils: modprobe foo'
    # internal/applets/procps/modutils: the name validates, the plan is
    # reported ("validated successfully"), then the CAP_SYS_MODULE gate fails.
    It 'validates the module then fails on the CAP_SYS_MODULE gate'
      When run modprobe foo
      The status should be failure
      The stderr should include 'modprobe: load of foo validated successfully, but inserting/removing kernel modules requires CAP_SYS_MODULE; this privileged step is intentionally not implemented in the hermetic build'
    End
  End
End

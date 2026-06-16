package ischroot

import "testing"

// TestIsFakeChroot drives every branch of isFakeChroot by toggling the
// FAKECHROOT, FAKECHROOT_BASE and LD_PRELOAD environment variables.
func TestIsFakeChroot(t *testing.T) {
	tests := []struct {
		name          string
		fakeChroot    string
		fakeChrootSet bool
		baseDir       string
		baseDirSet    bool
		ldPreload     string
		ldPreloadSet  bool
		want          bool
	}{
		{name: "not set", want: false},
		{name: "fakechroot not true", fakeChroot: "false", fakeChrootSet: true, want: false},
		{
			name: "no base dir", fakeChroot: "true", fakeChrootSet: true, want: false,
		},
		{
			name:       "base dir but no libfakechroot in preload",
			fakeChroot: "true", fakeChrootSet: true,
			baseDir: "/fake", baseDirSet: true,
			ldPreload: "/lib/other.so", ldPreloadSet: true,
			want: false,
		},
		{
			name:       "full fakechroot environment",
			fakeChroot: "true", fakeChrootSet: true,
			baseDir: "/fake", baseDirSet: true,
			ldPreload: "/usr/lib/libfakechroot.so", ldPreloadSet: true,
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			setOrUnset(t, "FAKECHROOT", tt.fakeChroot, tt.fakeChrootSet)
			setOrUnset(t, "FAKECHROOT_BASE", tt.baseDir, tt.baseDirSet)
			setOrUnset(t, "LD_PRELOAD", tt.ldPreload, tt.ldPreloadSet)

			if got := isFakeChroot(); got != tt.want {
				t.Errorf("isFakeChroot() = %v, want %v", got, tt.want)
			}
		})
	}
}

// setOrUnset sets env to value when set is true, otherwise unsets it; the
// original is restored after the test via t.Setenv semantics.
func setOrUnset(t *testing.T, key, value string, set bool) {
	t.Helper()
	if set {
		t.Setenv(key, value)
		return
	}
	// t.Setenv records the original value so it is restored on cleanup; setting
	// it to empty effectively neutralizes it for the duration of the test.
	t.Setenv(key, "")
}

// TestCanStatRootDir confirms the root directory is always statable, so the
// helper reports true on any supported platform.
func TestCanStatRootDir(t *testing.T) {
	if !canStatRootDir() {
		t.Error("canStatRootDir() = false, want true (root must be statable)")
	}
}

// TestIsNotJailReturnsBool exercises isNotJail without asserting a specific
// value: in the test environment /proc/1/root may or may not be statable. The
// goal is to drive the stat calls and ensure it does not panic.
func TestIsNotJailReturnsBool(t *testing.T) {
	_ = isNotJail()
}

// TestIsChrootReturnsValidCode checks that isChroot returns one of the defined
// exit codes.
func TestIsChrootReturnsValidCode(t *testing.T) {
	got := isChroot()
	switch got {
	case jail, notJail, notSuperUser:
		// valid
	default:
		t.Errorf("isChroot() = %d, want one of %d/%d/%d", got, jail, notJail, notSuperUser)
	}
}

// TestDetectFallbacks verifies that the -f/-t fallbacks only change the result
// when detection is undetermined (notSuperUser). When isChroot already gives a
// definite answer the fallbacks are ignored.
func TestDetectFallbacks(t *testing.T) {
	// Ensure no fakechroot environment interferes.
	t.Setenv("FAKECHROOT", "")
	base := isChroot()
	if base == notSuperUser {
		// When undetermined, -f maps to notJail and -t maps to jail.
		if got := detect(true, false); got != notJail {
			t.Errorf("detect(-f) = %d, want %d", got, notJail)
		}
		if got := detect(false, true); got != jail {
			t.Errorf("detect(-t) = %d, want %d", got, jail)
		}
	} else {
		// When determined, the fallbacks leave the result unchanged.
		if got := detect(true, false); got != base {
			t.Errorf("detect(-f) = %d, want unchanged %d", got, base)
		}
	}
}

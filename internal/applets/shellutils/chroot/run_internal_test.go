package chroot

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"testing"
)

// TestDecideExecCommand covers the three shapes: explicit command with args,
// explicit command alone, and the $SHELL fallback (interactive) when no command
// operand is given.
func TestDecideExecCommand(t *testing.T) {
	t.Run("explicit command with args", func(t *testing.T) {
		name, argv := decideExecCommand([]string{"/bin/echo", "hi", "there"})
		if name != "/bin/echo" {
			t.Errorf("name = %q, want /bin/echo", name)
		}
		if !reflect.DeepEqual(argv, []string{"hi", "there"}) {
			t.Errorf("argv = %v, want [hi there]", argv)
		}
	})

	t.Run("explicit command alone", func(t *testing.T) {
		name, argv := decideExecCommand([]string{"/bin/sh"})
		if name != "/bin/sh" {
			t.Errorf("name = %q, want /bin/sh", name)
		}
		if len(argv) != 0 {
			t.Errorf("argv = %v, want empty", argv)
		}
	})

	t.Run("shell fallback honors $SHELL", func(t *testing.T) {
		t.Setenv("SHELL", "/usr/bin/fish")
		name, argv := decideExecCommand(nil)
		if name != "/usr/bin/fish" {
			t.Errorf("name = %q, want /usr/bin/fish", name)
		}
		if !reflect.DeepEqual(argv, []string{"-i"}) {
			t.Errorf("argv = %v, want [-i]", argv)
		}
	})

	t.Run("shell fallback defaults to /bin/sh", func(t *testing.T) {
		t.Setenv("SHELL", "")
		name, argv := decideExecCommand(nil)
		if name != "/bin/sh" {
			t.Errorf("name = %q, want /bin/sh", name)
		}
		if !reflect.DeepEqual(argv, []string{"-i"}) {
			t.Errorf("argv = %v, want [-i]", argv)
		}
	})
}

// TestReason maps the recognized failure kinds to their GNU-style messages and
// falls back to the raw error text otherwise.
func TestReason(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"not exist", &fs.PathError{Op: "chroot", Path: "/x", Err: fs.ErrNotExist}, "No such file or directory"},
		{"permission", &fs.PathError{Op: "chroot", Path: "/x", Err: fs.ErrPermission}, "Operation not permitted"},
		{"other", errors.New("some other failure"), "some other failure"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := reason(tc.err); got != tc.want {
				t.Errorf("reason(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// TestApplySetgroupsError verifies apply surfaces a setgroups failure and stops
// before dropping the gid/uid, so a failed supplementary-group set never leaves
// the process with a partially-applied identity.
func TestApplySetgroupsError(t *testing.T) {
	origGroups, origGid, origUid := setgroups, setgid, setuid
	t.Cleanup(func() { setgroups, setgid, setuid = origGroups, origGid, origUid })

	gidCalled, uidCalled := false, false
	setgroups = func([]int) error { return os.ErrPermission }
	setgid = func(int) error { gidCalled = true; return nil }
	setuid = func(int) error { uidCalled = true; return nil }

	id := identity{uid: 1000, gid: 1000, groups: []int{10}}
	err := id.apply()
	if err == nil {
		t.Fatal("apply must fail when setgroups fails")
	}
	if gidCalled || uidCalled {
		t.Errorf("setgid/setuid must not run after a setgroups failure (gid=%v uid=%v)", gidCalled, uidCalled)
	}
}

// TestApplySetgidError verifies a setgid failure aborts before setuid.
func TestApplySetgidError(t *testing.T) {
	origGroups, origGid, origUid := setgroups, setgid, setuid
	t.Cleanup(func() { setgroups, setgid, setuid = origGroups, origGid, origUid })

	uidCalled := false
	setgroups = func([]int) error { return nil }
	setgid = func(int) error { return os.ErrPermission }
	setuid = func(int) error { uidCalled = true; return nil }

	id := identity{uid: 1000, gid: 1000, groups: nil}
	if err := id.apply(); err == nil {
		t.Fatal("apply must fail when setgid fails")
	}
	if uidCalled {
		t.Error("setuid must not run after a setgid failure")
	}
}

// TestApplySetuidError verifies a setuid failure is surfaced.
func TestApplySetuidError(t *testing.T) {
	origGroups, origGid, origUid := setgroups, setgid, setuid
	t.Cleanup(func() { setgroups, setgid, setuid = origGroups, origGid, origUid })

	setgroups = func([]int) error { return nil }
	setgid = func(int) error { return nil }
	setuid = func(int) error { return os.ErrPermission }

	id := identity{uid: 1000, gid: 1000, groups: nil}
	if err := id.apply(); err == nil {
		t.Fatal("apply must fail when setuid fails")
	}
}

package chroot

import (
	"reflect"
	"strings"
	"testing"
)

// jailPasswd and jailGroup model a jail whose account databases diverge from a
// typical host: "alice" is uid 1000 with primary gid 1000, "svc" is uid 4242,
// and the group names/ids do not line up with the host's defaults.
const (
	jailPasswd = `root:x:0:0:root:/root:/bin/sh
alice:x:1000:1000:Alice:/home/alice:/bin/sh
svc:x:4242:777:Service:/srv:/usr/sbin/nologin
# a comment line

malformed:x:notanumber:1:::
`
	jailGroup = `root:x:0:
devs:x:1000:alice
ops:x:777:svc
extra:x:9001:alice,svc
# comment
broken:x:notanumber:
`
)

func TestParsePasswd(t *testing.T) {
	t.Parallel()
	got := parsePasswd(strings.NewReader(jailPasswd))
	if e, ok := got["alice"]; !ok || e.uid != 1000 || e.gid != 1000 {
		t.Errorf("alice = %+v ok=%v, want uid=1000 gid=1000", e, ok)
	}
	if e, ok := got["svc"]; !ok || e.uid != 4242 || e.gid != 777 {
		t.Errorf("svc = %+v ok=%v, want uid=4242 gid=777", e, ok)
	}
	if _, ok := got["malformed"]; ok {
		t.Errorf("malformed line should be skipped")
	}
	if _, ok := got["#"]; ok {
		t.Errorf("comment line should not be parsed")
	}
}

func TestParsePasswdNil(t *testing.T) {
	t.Parallel()
	if got := parsePasswd(nil); len(got) != 0 {
		t.Errorf("parsePasswd(nil) = %v, want empty", got)
	}
}

func TestParseGroup(t *testing.T) {
	t.Parallel()
	got := parseGroup(strings.NewReader(jailGroup))
	for name, want := range map[string]int{"root": 0, "devs": 1000, "ops": 777, "extra": 9001} {
		if gid, ok := got[name]; !ok || gid != want {
			t.Errorf("group %q = %d ok=%v, want %d", name, gid, ok, want)
		}
	}
	if _, ok := got["broken"]; ok {
		t.Errorf("broken line should be skipped")
	}
}

func TestResolveIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		userSpec  string
		groupsCSV string
		want      identity
		wantErr   bool
	}{
		{
			name:     "user name resolves uid and primary gid from jail",
			userSpec: "alice",
			want:     identity{uid: 1000, gid: 1000, groups: []int{}},
		},
		{
			name:     "user:group names resolve against jail databases",
			userSpec: "alice:ops",
			want:     identity{uid: 1000, gid: 777, groups: []int{}},
		},
		{
			name:     "svc uses divergent primary gid 777",
			userSpec: "svc",
			want:     identity{uid: 4242, gid: 777, groups: []int{}},
		},
		{
			name:     "numeric user with no passwd entry defaults gid to uid",
			userSpec: "5000",
			want:     identity{uid: 5000, gid: 5000, groups: []int{}},
		},
		{
			name:     "numeric user and numeric group",
			userSpec: "5000:6000",
			want:     identity{uid: 5000, gid: 6000, groups: []int{}},
		},
		{
			name:      "userspec with supplementary groups",
			userSpec:  "alice:devs",
			groupsCSV: "ops,extra",
			want:      identity{uid: 1000, gid: 1000, groups: []int{777, 9001}},
		},
		{
			name:      "numeric supplementary groups",
			userSpec:  "alice",
			groupsCSV: "10,20",
			want:      identity{uid: 1000, gid: 1000, groups: []int{10, 20}},
		},
		{
			name:     "unknown user name is rejected",
			userSpec: "nobodyhere",
			wantErr:  true,
		},
		{
			name:     "unknown group name is rejected",
			userSpec: "alice:nogroup",
			wantErr:  true,
		},
		{
			name:     "empty user part is rejected",
			userSpec: ":devs",
			wantErr:  true,
		},
		{
			name:     "trailing colon (empty group) is rejected",
			userSpec: "alice:",
			wantErr:  true,
		},
		{
			name:      "groups without userspec is rejected",
			groupsCSV: "devs",
			wantErr:   true,
		},
		{
			name:      "unknown supplementary group is rejected",
			userSpec:  "alice",
			groupsCSV: "ops,nope",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveIdentity(
				tt.userSpec, tt.groupsCSV,
				strings.NewReader(jailPasswd), strings.NewReader(jailGroup),
			)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveIdentity(%q,%q) = %+v, want error", tt.userSpec, tt.groupsCSV, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveIdentity(%q,%q) unexpected error: %v", tt.userSpec, tt.groupsCSV, err)
			}
			if got.uid != tt.want.uid || got.gid != tt.want.gid || !reflect.DeepEqual(got.groups, tt.want.groups) {
				t.Errorf("resolveIdentity(%q,%q) = %+v, want %+v", tt.userSpec, tt.groupsCSV, got, tt.want)
			}
		})
	}
}

// TestResolveIdentityNilDatabases verifies that with no jail passwd/group files
// only numeric IDs resolve and name lookups fail deterministically.
func TestResolveIdentityNilDatabases(t *testing.T) {
	t.Parallel()
	got, err := resolveIdentity("1234:5678", "", nil, nil)
	if err != nil {
		t.Fatalf("numeric userspec with nil databases: %v", err)
	}
	if got.uid != 1234 || got.gid != 5678 {
		t.Errorf("got %+v, want uid=1234 gid=5678", got)
	}
	if _, err := resolveIdentity("alice", "", nil, nil); err == nil {
		t.Errorf("name lookup with nil databases should fail")
	}
}

// TestApplyOrder verifies apply drops privileges in the order
// setgroups -> setgid -> setuid, since after setuid the process can no longer
// change its gid or groups.
func TestApplyOrder(t *testing.T) {
	origGroups, origGid, origUid := setgroups, setgid, setuid
	t.Cleanup(func() { setgroups, setgid, setuid = origGroups, origGid, origUid })

	var order []string
	setgroups = func(g []int) error { order = append(order, "groups"); return nil }
	setgid = func(g int) error { order = append(order, "gid"); return nil }
	setuid = func(u int) error { order = append(order, "uid"); return nil }

	id := identity{uid: 1000, gid: 1000, groups: []int{10}}
	if err := id.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{"groups", "gid", "uid"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("apply order = %v, want %v", order, want)
	}
}

func TestApplyNilGroupsSkipsSetgroups(t *testing.T) {
	origGroups, origGid, origUid := setgroups, setgid, setuid
	t.Cleanup(func() { setgroups, setgid, setuid = origGroups, origGid, origUid })

	called := false
	setgroups = func(g []int) error { called = true; return nil }
	setgid = func(g int) error { return nil }
	setuid = func(u int) error { return nil }

	id := identity{uid: 1000, gid: 1000, groups: nil}
	if err := id.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if called {
		t.Errorf("setgroups should not be called when groups is nil")
	}
}

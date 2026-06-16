package ipcmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestDumpDispatch exercises the shared backend dispatch directly, independent
// of any command surface, so each object renders through the one routine the
// whole ip family funnels into.
func TestDumpDispatch(t *testing.T) {
	defer SetSource(fixture())()

	tests := []struct {
		name   string
		obj    object
		target []string
		want   string
	}{
		{name: "link", obj: objLink, want: "2: eth0:"},
		{name: "link dev filter", obj: objLink, target: []string{"dev", "eth0"}, want: "eth0"},
		{name: "addr", obj: objAddr, want: "inet 192.168.1.10/24 scope global"},
		{name: "route", obj: objRoute, want: "default via 192.168.1.1 dev eth0 proto static"},
		{name: "neigh", obj: objNeigh, want: "192.168.1.1 dev eth0 lladdr 52:54:00:aa:bb:cc REACHABLE"},
		{name: "rule", obj: objRule, want: "from all lookup local"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			dump(&b, tt.obj, tt.target)
			if !strings.Contains(b.String(), tt.want) {
				t.Errorf("dump(%v) missing %q\ngot:\n%s", tt.name, tt.want, b.String())
			}
		})
	}
}

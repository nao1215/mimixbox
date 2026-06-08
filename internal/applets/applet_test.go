// mimixbox/internal/applets/applet_test.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package applets

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/pidof"
)

// TestPidofIsRegistered guards against the regression where pidof had a real
// implementation and unit tests in the tree but was never wired into the applet
// registry. Without registration, "mimixbox pidof", "--list" and
// "--full-install" silently omit it, and the ShellSpec suite ends up exercising
// the host pidof instead of MimixBox's own. See GitHub issue #265.
func TestPidofIsRegistered(t *testing.T) {
	t.Parallel()

	name := pidof.New().Name()
	if !HasApplet(name) {
		t.Fatalf("applet %q is implemented but not registered in Applets", name)
	}

	if got, want := Applets[name].Desc, pidof.New().Synopsis(); got != want {
		t.Errorf("registered description drifted from the command synopsis:\n got: %q\nwant: %q", got, want)
	}
}

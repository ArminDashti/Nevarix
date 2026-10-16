package lifecycle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidPort(t *testing.T) {
	for _, ok := range []int{1, 8181, 65535} {
		if err := validPort(ok); err != nil {
			t.Fatalf("%d rejected: %v", ok, err)
		}
	}
	for _, bad := range []int{-1, 0, 65536} {
		if validPort(bad) == nil {
			t.Fatalf("%d accepted", bad)
		}
	}
}

// Staging must sit outside the app base, otherwise --remove-data wipes it mid-build.
func TestStagingOutsideAppBase(t *testing.T) {
	p, err := App{Name: "core"}.paths()
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(p.staging, p.base+string(filepath.Separator)) {
		t.Fatalf("staging %s inside base %s", p.staging, p.base)
	}
}

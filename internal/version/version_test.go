package version

import "testing"

func TestCompare(t *testing.T) {
	// Each version is lower than the next, per the semver.org precedence example.
	ordered := []string{
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0", "10.0.0",
	}
	for i := 0; i < len(ordered)-1; i++ {
		a, b := ordered[i], ordered[i+1]
		if Compare(a, b) != -1 || Compare(b, a) != 1 {
			t.Errorf("expected %s < %s", a, b)
		}
	}
	if Compare("v0.1.0", "0.1.0+build.5") != 0 {
		t.Error("prefix and build metadata should be ignored")
	}
}

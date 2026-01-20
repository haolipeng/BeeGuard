package host

import (
	"testing"
)

func TestPlatformInfo(t *testing.T) {
	if Platform == "" {
		t.Log("Platform information may not be available on this system")
	} else {
		t.Logf("Platform: %s", Platform)
	}

	if PlatformFamily == "" {
		t.Log("PlatformFamily may not be available on this system")
	} else {
		t.Logf("PlatformFamily: %s", PlatformFamily)
	}

	if PlatformVersion == "" {
		t.Log("PlatformVersion may not be available on this system")
	} else {
		t.Logf("PlatformVersion: %s", PlatformVersion)
	}

	if KernelVersion == "" {
		t.Error("KernelVersion should be available")
	} else {
		t.Logf("KernelVersion: %s", KernelVersion)
	}

	if Arch == "" {
		t.Error("Arch should be available")
	} else {
		t.Logf("Arch: %s", Arch)
	}
}

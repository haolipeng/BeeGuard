package host

import (
	"testing"
)

func TestRefreshHost(t *testing.T) {
	RefreshHost()

	name := Name.Load()
	if name == nil || name.(string) == "" {
		t.Error("Hostname should not be empty")
	} else {
		t.Logf("Hostname: %s", name.(string))
	}

	ipv4 := IPv4.Load()
	if ipv4 == nil {
		t.Error("IPv4 list should not be nil")
	} else {
		ipv4List := ipv4.([]string)
		t.Logf("IPv4 addresses: %v", ipv4List)
		if len(ipv4List) > 10 {
			t.Errorf("IPv4 list should not exceed 10, got %d", len(ipv4List))
		}
	}
}

func TestHostname(t *testing.T) {
	name := Name.Load()
	if name == nil {
		t.Error("Hostname should be initialized")
		return
	}
	hostname := name.(string)
	if hostname == "" {
		t.Error("Hostname should not be empty")
	}
	t.Logf("Hostname: %s", hostname)
}

func TestIPv4Collection(t *testing.T) {
	ipv4 := IPv4.Load()
	if ipv4 == nil {
		t.Error("IPv4 list should be initialized")
		return
	}
	ipv4List := ipv4.([]string)
	t.Logf("IPv4 addresses count: %d", len(ipv4List))
	for _, ip := range ipv4List {
		t.Logf("  - %s", ip)
	}
}

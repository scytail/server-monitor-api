package services

import (
	"encoding/json"
	"reflect"
	"testing"

	systemd "github.com/coreos/go-systemd/v22/dbus"
)

func TestEncodeServices(t *testing.T) {
	units := []systemd.UnitStatus{
		{
			Name:        "zeta.service",
			LoadState:   "loaded",
			ActiveState: "inactive",
			SubState:    "dead",
			Description: "Zeta service",
		},
		{
			Name:        "alpha.service",
			LoadState:   "loaded",
			ActiveState: "active",
			SubState:    "running",
			Description: "Alpha service",
		},
		{Name: "timer.timer", LoadState: "loaded"},
	}

	data, err := encodeServices(units)
	if err != nil {
		t.Fatalf("encodeServices returned error: %v", err)
	}

	var services []Service
	if err := json.Unmarshal([]byte(data), &services); err != nil {
		t.Fatalf("invalid JSON returned: %v", err)
	}

	want := []Service{
		{
			Name:        "alpha.service",
			LoadStatus:  "loaded",
			Active:      true,
			Status:      "running",
			Description: "Alpha service",
		},
		{
			Name:        "zeta.service",
			LoadStatus:  "loaded",
			Active:      false,
			Status:      "dead",
			Description: "Zeta service",
		},
	}
	if !reflect.DeepEqual(services, want) {
		t.Errorf("encodeServices() = %#v, want %#v", services, want)
	}
}

func TestEncodeServicesEmptyList(t *testing.T) {
	data, err := encodeServices(nil)
	if err != nil {
		t.Fatalf("encodeServices returned error: %v", err)
	}
	if data != "[]" {
		t.Errorf("encodeServices(nil) = %s, want []", data)
	}
}

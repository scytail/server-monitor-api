package services

import (
	"encoding/json"
	"reflect"
	"testing"

	systemd "github.com/coreos/go-systemd/v22/dbus"
)

func TestEncodeServices(t *testing.T) {
	input := []Service{
		{
			Name:        "zeta.timer",
			LoadStatus:  "loaded",
			Active:      false,
			Status:      "dead",
			Description: "Zeta timer",
		},
		{
			Name:        "alpha.service",
			LoadStatus:  "loaded",
			Active:      true,
			Status:      "running",
			Description: "Alpha service",
		},
	}

	data, err := encodeServices(input)
	if err != nil {
		t.Fatalf("encodeServices returned error: %v", err)
	}

	var services []Service
	if err := json.Unmarshal([]byte(data), &services); err != nil {
		t.Fatalf("invalid JSON returned: %v", err)
	}

	if !reflect.DeepEqual(services, input) {
		t.Errorf("encodeServices() = %#v, want %#v", services, input)
	}
}

func TestEncodeServicesEmptyList(t *testing.T) {
	data, err := encodeServices(nil)
	if err != nil {
		t.Fatalf("encodeServices returned error: %v", err)
	}
	if data != "null" {
		t.Errorf("encodeServices(nil) = %s, want null", data)
	}
}

func TestFilterServices(t *testing.T) {
	units := []systemd.UnitStatus{
		{Name: "zeta.service", LoadState: "loaded"},
		{Name: "alpha.service", LoadState: "loaded"},
		{Name: "timer.timer", LoadState: "loaded"},
	}

	tests := []struct {
		name             string
		selectedServices []string
		want             []Service
	}{
		{
			name:             "selected services",
			selectedServices: []string{"zeta.service"},
			want:             []Service{{Name: "zeta.service", LoadStatus: "loaded"}},
		},
		{
			name:             "no selected services",
			selectedServices: nil,
			want: []Service{
				{Name: "alpha.service", LoadStatus: "loaded"},
				{Name: "zeta.service", LoadStatus: "loaded"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterServices(units, tt.selectedServices)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterServices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

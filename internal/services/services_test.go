package services

import (
	"reflect"
	"testing"

	systemd "github.com/coreos/go-systemd/v22/dbus"
)

func TestConvertToServiceList(t *testing.T) {
	units := []systemd.UnitStatus{
		{
			Name:        "zeta.service",
			LoadState:   "loaded",
			ActiveState: "active",
			SubState:    "running",
			Description: "Zeta service",
		},
		{
			Name:        "alpha.service",
			LoadState:   "loaded",
			ActiveState: "inactive",
			SubState:    "dead",
			Description: "Alpha service",
		},
		{Name: "timer.timer", LoadState: "loaded"},
	}

	got := convertToServiceList(units)
	want := []Service{
		{
			Name:        "alpha.service",
			LoadStatus:  "loaded",
			Active:      false,
			Status:      "dead",
			Description: "Alpha service",
		},
		{
			Name:        "zeta.service",
			LoadStatus:  "loaded",
			Active:      true,
			Status:      "running",
			Description: "Zeta service",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("convertToServiceList() = %#v, want %#v", got, want)
	}
}

func TestConvertToServiceListEmpty(t *testing.T) {
	got := convertToServiceList(nil)
	if !reflect.DeepEqual(got, []Service{}) {
		t.Errorf("convertToServiceList(nil) = %#v, want an empty service list", got)
	}
}

func TestFilterServiceNames(t *testing.T) {
	input := []Service{
		{Name: "alpha.service"},
		{Name: "zeta.service"},
	}

	tests := []struct {
		name  string
		names []string
		want  []Service
	}{
		{
			name:  "selected names",
			names: []string{"zeta.service"},
			want:  []Service{{Name: "zeta.service"}},
		},
		{
			name:  "empty names",
			names: []string{},
			want:  []Service{},
		},
		{
			name:  "nil names",
			names: nil,
			want:  []Service{},
		},
		{
			name:  "unknown names",
			names: []string{"missing.service"},
			want:  []Service{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterServices(input, tt.names)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterServices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

package services

import (
	"os"
	"path/filepath"
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

func TestFilterBlacklistedServices(t *testing.T) {
	input := []Service{
		{Name: "ssh.service"},
		{Name: "docker.service"},
		{Name: "docker.socket"},
		{Name: "cron.service"},
	}

	tests := []struct {
		name     string
		patterns []string
		want     []Service
	}{
		{
			name:     "exact service name",
			patterns: []string{"ssh.service"},
			want:     []Service{{Name: "docker.service"}, {Name: "docker.socket"}, {Name: "cron.service"}},
		},
		{
			name:     "wildcards before and after",
			patterns: []string{"*docker*"},
			want:     []Service{{Name: "ssh.service"}, {Name: "cron.service"}},
		},
		{
			name:     "wildcard matches service suffix",
			patterns: []string{"*.service"},
			want:     []Service{{Name: "docker.socket"}},
		},
		{
			name:     "blank patterns are ignored",
			patterns: []string{" ", ""},
			want:     input,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterBlacklistedServices(input, tt.patterns)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterBlacklistedServices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestFilterUnitServicesBeforeRequestedNameFiltering(t *testing.T) {
	units := []systemd.UnitStatus{
		{Name: "docker.service", LoadState: "loaded"},
		{Name: "cron.service", LoadState: "loaded"},
	}

	allServices := filterUnitServices(units, []string{"*docker*"})
	got := FilterServices(allServices, []string{"docker.service", "cron.service"})
	want := []Service{{Name: "cron.service", LoadStatus: "loaded"}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FilterServices() after blacklist = %#v, want %#v", got, want)
	}
}

func TestLoadBlacklist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blacklist")
	content := "# comment\nssh.service\n  # indented comment\n*docker*\n \t\nfoo#bar\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write blacklist: %v", err)
	}

	got, found, err := LoadBlacklist(path)
	if err != nil {
		t.Fatalf("LoadBlacklist() error = %v, want nil", err)
	}
	if !found {
		t.Fatal("LoadBlacklist() found = false, want true")
	}
	want := []string{"ssh.service", "*docker*", "foo#bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadBlacklist() = %#v, want %#v", got, want)
	}
}

func TestLoadBlacklistMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-blacklist")

	got, found, err := LoadBlacklist(path)
	if err != nil {
		t.Fatalf("LoadBlacklist() error = %v, want nil", err)
	}
	if found {
		t.Fatal("LoadBlacklist() found = true, want false")
	}
	if !reflect.DeepEqual(got, []string{}) {
		t.Errorf("LoadBlacklist() = %#v, want an empty blacklist", got)
	}
}

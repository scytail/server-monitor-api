package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	systemd "github.com/coreos/go-systemd/v22/dbus"
)

type Service struct {
	Name        string `json:"name"`
	LoadStatus  string `json:"loadStatus"`
	Active      bool   `json:"active"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func ListServices(ctx context.Context) ([]Service, error) {
	// connect to systemd
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect to systemd: %w", err)
	}
	defer conn.Close()

	// query all service units
	units, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query systemd units: %w", err)
	}

	// interpolate service units into a usable list
	return convertToServiceList(units), nil
}

func FilterServices(services []Service, names []string) []Service {
	selected := make(map[string]struct{}, len(names))
	for _, name := range names {
		selected[name] = struct{}{}
	}

	filtered := make([]Service, 0, len(names))
	for _, service := range services {
		if _, ok := selected[service.Name]; ok {
			filtered = append(filtered, service)
		}
	}

	return filtered
}

func convertToServiceList(units []systemd.UnitStatus) []Service {
	services := make([]Service, 0, len(units))
	for _, unit := range units {
		if !strings.HasSuffix(unit.Name, ".service") {
			continue
		}

		services = append(services, Service{
			Name:        unit.Name,
			LoadStatus:  unit.LoadState,
			Active:      unit.ActiveState == "active",
			Status:      unit.SubState,
			Description: unit.Description,
		})
	}

	// Sort alphabetically
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services
}

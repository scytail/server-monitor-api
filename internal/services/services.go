package services

import (
	"context"
	"encoding/json"
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

func ListServices(ctx context.Context) (string, error) {
	// connect to systemd
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return "[]", fmt.Errorf("connect to systemd: %w", err)
	}
	defer conn.Close()

	// query all service units
	units, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return "[]", fmt.Errorf("query systemd units: %w", err)
	}

	// encode the service units as JSON
	return encodeServices(units)
}

func encodeServices(units []systemd.UnitStatus) (string, error) {
	// Build a list of services from the units
	services := make([]Service, 0, len(units))
	for _, unit := range units {
		if strings.HasSuffix(unit.Name, ".service") {
			services = append(services, Service{
				Name:        unit.Name,
				LoadStatus:  unit.LoadState,
				Active:      unit.ActiveState == "active",
				Status:      unit.SubState,
				Description: unit.Description,
			})
		}
	}

	// Sort the services by name
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	// Serialize the service list
	data, err := json.Marshal(services)
	if err != nil {
		return "[]", fmt.Errorf("encode services as JSON: %w", err)
	}

	return string(data), nil
}

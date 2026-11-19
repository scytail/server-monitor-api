package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
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

func GetAllServices(ctx context.Context, blacklist []string) ([]Service, error) {
	// connect to systemd
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect to systemd: %w", err)
	}
	defer conn.Close()

	// query all systemd units
	units, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query systemd units: %w", err)
	}

	return filterUnitServices(units, blacklist), nil
}

func filterUnitServices(units []systemd.UnitStatus, blacklist []string) []Service {
	return FilterBlacklistedServices(convertToServiceList(units), blacklist)
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

func LoadBlacklist(path string) ([]string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, false, nil
		}
		return nil, false, fmt.Errorf("open blacklist file %q: %w", path, err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, false, fmt.Errorf("read blacklist file %q: %w", path, err)
	}

	return patterns, true, nil
}

func FilterBlacklistedServices(services []Service, patterns []string) []Service {
	matchers := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		expression := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*") + "$"
		matchers = append(matchers, regexp.MustCompile(expression))
	}

	filtered := make([]Service, 0, len(services))
	for _, service := range services {
		blacklisted := false
		for _, matcher := range matchers {
			if matcher.MatchString(service.Name) {
				blacklisted = true
				break
			}
		}
		if !blacklisted {
			filtered = append(filtered, service)
		}
	}

	return filtered
}

func convertToServiceList(units []systemd.UnitStatus) []Service {
	services := make([]Service, 0, len(units))
	for _, unit := range units {
		// Skip anything that doesn't have a "service" suffix
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

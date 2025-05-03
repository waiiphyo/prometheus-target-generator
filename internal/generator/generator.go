package generator

import (
	"fmt"
	"strings"
)

// TargetConfig defines the Prometheus target format.
type TargetConfig struct {
	Labels  map[string]string `json:"labels"`
	Targets []string          `json:"targets"`
}

// GenerateTargets processes IP addresses into a Prometheus target format.
func GenerateTargets(groupName string, ipData string) ([]TargetConfig, error) {
	// Split the IP addresses by commas
	ips := strings.Split(ipData, ",")
	var targetConfigs []TargetConfig

	// Loop over the IPs to create a Prometheus target configuration for each
	for i, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// Generate a sequential hostname like group-name-01, group-name-02
		hostname := fmt.Sprintf("%s-%02d", groupName, i+1)

		// Create target configuration
		targetConfig := TargetConfig{
			Labels: map[string]string{
				"job":      groupName,
				"instance": hostname,
			},
			Targets: []string{fmt.Sprintf("%s:9100", ip)},
		}

		targetConfigs = append(targetConfigs, targetConfig)
	}

	return targetConfigs, nil
}

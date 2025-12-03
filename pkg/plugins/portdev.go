package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&PortDevPlugin{})
}

// PortDevPlugin extracts port device configuration
type PortDevPlugin struct{}

func (p *PortDevPlugin) Name() string {
	return "portdev"
}

func (p *PortDevPlugin) Description() string {
	return "Port devices configuration - SYSTEM\\ControlSet\\Control\\COM Name Arbiter"
}

func (p *PortDevPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *PortDevPlugin) Run(hive *regf.Hive) error {
	controlSetName, err := findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	portPath := fmt.Sprintf("%s\\Control\\COM Name Arbiter", controlSetName)
	portKey, err := hive.GetKey(portPath)
	if err != nil {
		return fmt.Errorf("port device key not found: %w", err)
	}

	fmt.Println("Port Devices:")
	fmt.Println(strings.Repeat("=", 80))

	for _, val := range portKey.Values() {
		if strings.HasPrefix(val.Name(), "ComDB") {
			continue
		}
		fmt.Printf("%-20s: %s\n", val.Name(), GetValueString(val))
	}

	// Try Devices subkey
	if devices, err := getSubkey(portKey, "Devices"); err == nil {
		fmt.Println("\nConnected Devices:")
		for _, val := range devices.Values() {
			fmt.Printf("%-20s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}

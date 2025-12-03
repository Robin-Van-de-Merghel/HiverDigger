package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&USBSTORPlugin{})
}

// USBSTORPlugin extracts USB storage device history from SYSTEM hive
type USBSTORPlugin struct{}

func (p *USBSTORPlugin) Name() string {
	return "usbstor"
}

func (p *USBSTORPlugin) Description() string {
	return "USB storage device history - SYSTEM\\ControlSet\\Enum\\USBSTOR"
}

func (p *USBSTORPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *USBSTORPlugin) Run(hive *regf.Hive) error {
	controlSetName, err := findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	usbstorPath := fmt.Sprintf("%s\\Enum\\USBSTOR", controlSetName)
	usbstorKey, err := hive.GetKey(usbstorPath)
	if err != nil {
		return fmt.Errorf("USBSTOR key not found: %w", err)
	}

	fmt.Println("USB Storage Devices:")
	fmt.Println(strings.Repeat("=", 80))

	for _, deviceType := range usbstorKey.Subkeys() {
		for _, instance := range deviceType.Subkeys() {
			fmt.Printf("\nDevice: %s\n", deviceType.Name())
			fmt.Printf("  Serial: %s\n", instance.Name())
			
			for _, val := range instance.Values() {
				if val.Name() == "FriendlyName" || val.Name() == "DeviceDesc" ||
					val.Name() == "ParentIdPrefix" {
					fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
				}
			}
		}
	}

	return nil
}

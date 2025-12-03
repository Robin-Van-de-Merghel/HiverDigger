package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&USBSTOR2Plugin{})
}

// USBSTOR2Plugin provides enhanced USB storage information
type USBSTOR2Plugin struct{}

func (p *USBSTOR2Plugin) Name() string {
	return "usbstor2"
}

func (p *USBSTOR2Plugin) Description() string {
	return "Enhanced USB storage info with timestamps"
}

func (p *USBSTOR2Plugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *USBSTOR2Plugin) Run(hive *regf.Hive) error {
	controlSetName, err := findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	usbstorPath := fmt.Sprintf("%s\\Enum\\USBSTOR", controlSetName)
	usbstorKey, err := hive.GetKey(usbstorPath)
	if err != nil {
		return fmt.Errorf("USBSTOR key not found: %w", err)
	}

	fmt.Println("USB Storage Devices (Enhanced):")
	fmt.Println(strings.Repeat("=", 80))

	for _, deviceType := range usbstorKey.Subkeys() {
		for _, instance := range deviceType.Subkeys() {
			fmt.Printf("\n[%s]\n", instance.Timestamp().Format("2006-01-02 15:04:05"))
			fmt.Printf("Device: %s\n", deviceType.Name())
			fmt.Printf("Instance: %s\n", instance.Name())

			if props, err := getSubkey(instance, "Properties"); err == nil {
				for _, guidKey := range props.Subkeys() {
					for _, propKey := range guidKey.Subkeys() {
						for _, val := range propKey.Values() {
							if val.Name() == "Data" {
								fmt.Printf("  Property: %s\n", GetValueString(val))
							}
						}
					}
				}
			}
		}
	}

	return nil
}

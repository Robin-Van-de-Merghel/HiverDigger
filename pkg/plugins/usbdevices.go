package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&USBDevicesPlugin{})
}

// USBDevicesPlugin lists USB devices from SYSTEM hive.
type USBDevicesPlugin struct{}

func (p *USBDevicesPlugin) Name() string {
	return "usbdevices"
}

func (p *USBDevicesPlugin) Description() string {
	return "List USB devices from SYSTEM hive"
}

func (p *USBDevicesPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *USBDevicesPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Enum\\USBSTOR",
		"ControlSet001\\Enum\\USB",
	}

	fmt.Println("USB Devices")
	fmt.Println("===========")
	fmt.Println()

	for _, basePath := range paths {
		key, err := hive.GetKey(basePath)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", basePath)
		p.listUSBDevices(key, 0)
		fmt.Println()
	}

	return nil
}

func (p *USBDevicesPlugin) listUSBDevices(key *regf.Key, depth int) {
	indent := strings.Repeat("  ", depth)

	// Check for FriendlyName or DeviceDesc
	friendlyName := ""
	for _, val := range key.Values() {
		name := val.Name()
		if strings.EqualFold(name, "FriendlyName") || strings.EqualFold(name, "DeviceDesc") {
			friendlyName = GetValueString(val)
			break
		}
	}

	if friendlyName != "" {
		fmt.Printf("%s%s: %s\n", indent, key.Name(), friendlyName)
	} else if depth > 0 {
		fmt.Printf("%s%s\n", indent, key.Name())
	}

	// Recurse into subkeys
	for _, subkey := range key.Subkeys() {
		p.listUSBDevices(subkey, depth+1)
	}
}

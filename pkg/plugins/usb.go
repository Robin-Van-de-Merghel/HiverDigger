package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&USBPlugin{})
}

// USBPlugin extracts detailed USB device enumeration from SYSTEM hive
type USBPlugin struct{}

func (p *USBPlugin) Name() string {
	return "usb"
}

func (p *USBPlugin) Description() string {
	return "USB device enumeration (detailed) - SYSTEM\\ControlSet\\Enum\\USB"
}

func (p *USBPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *USBPlugin) Run(hive *regf.Hive) error {
	controlSetName, err := findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	usbPath := fmt.Sprintf("%s\\Enum\\USB", controlSetName)
	usbKey, err := hive.GetKey(usbPath)
	if err != nil {
		return fmt.Errorf("USB key not found: %w", err)
	}

	fmt.Println("USB Devices (Enumerated):")
	fmt.Println(strings.Repeat("=", 80))

	for _, deviceClass := range usbKey.Subkeys() {
		for _, device := range deviceClass.Subkeys() {
			fmt.Printf("\nDevice: %s\\%s\n", deviceClass.Name(), device.Name())

			for _, val := range device.Values() {
				if val.Name() == "DeviceDesc" || val.Name() == "FriendlyName" ||
					val.Name() == "Mfg" || val.Name() == "Service" {
					fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
				}
			}
		}
	}

	return nil
}

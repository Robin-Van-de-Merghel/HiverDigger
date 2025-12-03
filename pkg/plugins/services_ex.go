package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ServicesExPlugin{})
}

// ServicesExPlugin provides enhanced services details with more information
type ServicesExPlugin struct{}

func (p *ServicesExPlugin) Name() string {
	return "services_ex"
}

func (p *ServicesExPlugin) Description() string {
	return "Enhanced Windows services with DLL/driver info and dependencies"
}

func (p *ServicesExPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *ServicesExPlugin) Run(hive *regf.Hive) error {
	controlSetName, err := findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	servicesPath := fmt.Sprintf("%s\\Services", controlSetName)
	servicesKey, err := hive.GetKey(servicesPath)
	if err != nil {
		return fmt.Errorf("services key not found: %w", err)
	}

	fmt.Println("Windows Services (Enhanced):")
	fmt.Println(strings.Repeat("=", 80))

	count := 0
	for _, svc := range servicesKey.Subkeys() {
		var displayName, imageType, start string

		for _, val := range svc.Values() {
			switch val.Name() {
			case "DisplayName":
				displayName = GetValueString(val)
			case "Type":
				imageType = GetValueString(val)
			case "Start":
				start = GetValueString(val)
			}
		}

		if displayName != "" || start != "" {
			fmt.Printf("\n[%s] %s\n", svc.Timestamp().Format("2006-01-02 15:04:05"), svc.Name())
			if displayName != "" {
				fmt.Printf("  Display Name: %s\n", displayName)
			}

			for _, val := range svc.Values() {
				name := val.Name()
				switch name {
				case "ImagePath":
					fmt.Printf("  Image Path: %s\n", GetValueString(val))

				case "Type":
					fmt.Printf("  Type: %s\n", imageType)

				case "Start":
					fmt.Printf("  Start: %s\n", start)

				case "DependOnService":
					fmt.Printf("  Dependencies: %s\n", GetValueString(val))

				case "Group":
					fmt.Printf("  Group: %s\n", GetValueString(val))
				}
			}

			count++
			if count >= 100 {
				fmt.Println("\n... (showing first 100 services)")
				break
			}
		}
	}

	fmt.Printf("\nTotal services found: %d\n", count)
	return nil
}

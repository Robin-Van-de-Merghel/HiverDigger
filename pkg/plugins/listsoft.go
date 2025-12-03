package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ListSoftPlugin{})
}

// ListSoftPlugin lists installed software from the SOFTWARE hive.
// Based on RegRipper's listsoft.pl plugin.
type ListSoftPlugin struct{}

func (p *ListSoftPlugin) Name() string {
	return "listsoft"
}

func (p *ListSoftPlugin) Description() string {
	return "List installed software from SOFTWARE hive (Uninstall keys)"
}

func (p *ListSoftPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *ListSoftPlugin) Run(hive *regf.Hive) error {
	fmt.Println("Installed Software")
	fmt.Println("==================")
	fmt.Println()

	// Check both 32-bit and 64-bit uninstall locations
	paths := []string{
		"Microsoft\\Windows\\CurrentVersion\\Uninstall",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
	}

	for _, path := range paths {
		if err := p.listUninstallKeys(hive, path); err != nil {
			// Continue to next path if this one fails
			continue
		}
	}

	return nil
}

func (p *ListSoftPlugin) listUninstallKeys(hive *regf.Hive, basePath string) error {
	key, err := hive.GetKey(basePath)
	if err != nil {
		return err
	}

	subkeys := key.Subkeys()
	for _, subkey := range subkeys {
		displayName := ""
		displayVersion := ""
		publisher := ""
		installDate := ""
		installLocation := ""

		for _, val := range subkey.Values() {
			name := val.Name()
			switch {
			case strings.EqualFold(name, "DisplayName"):
				displayName = GetValueString(val)
			case strings.EqualFold(name, "DisplayVersion"):
				displayVersion = GetValueString(val)
			case strings.EqualFold(name, "Publisher"):
				publisher = GetValueString(val)
			case strings.EqualFold(name, "InstallDate"):
				installDate = GetValueString(val)
			case strings.EqualFold(name, "InstallLocation"):
				installLocation = GetValueString(val)
			}
		}

		// Only display if we have a display name
		if displayName != "" {
			fmt.Printf("Software: %s\n", displayName)
			if displayVersion != "" {
				fmt.Printf("  Version: %s\n", displayVersion)
			}
			if publisher != "" {
				fmt.Printf("  Publisher: %s\n", publisher)
			}
			if installDate != "" {
				fmt.Printf("  Install Date: %s\n", installDate)
			}
			if installLocation != "" {
				fmt.Printf("  Install Location: %s\n", installLocation)
			}
			fmt.Println()
		}
	}

	return nil
}

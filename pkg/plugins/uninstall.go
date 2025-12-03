package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&UninstallPlugin{})
}

// UninstallPlugin lists programs from Uninstall registry keys (similar to listsoft but more comprehensive).
type UninstallPlugin struct{}

func (p *UninstallPlugin) Name() string {
	return "uninstall"
}

func (p *UninstallPlugin) Description() string {
	return "Comprehensive list of installed/uninstalled programs from SOFTWARE hive"
}

func (p *UninstallPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *UninstallPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft/Windows/CurrentVersion/Uninstall",
		"Wow6432Node/Microsoft/Windows/CurrentVersion/Uninstall",
	}

	fmt.Println("Uninstall Registry Entries:")
	fmt.Println(strings.Repeat("=", 80))

	count := 0
	for _, basePath := range paths {
		key, err := hive.GetKey(basePath)
		if err != nil {
			continue
		}

		fmt.Printf("\n%s:\n", basePath)

		for _, subkey := range key.Subkeys() {
			var displayName, displayVersion, publisher, installDate, uninstallString string

			for _, val := range subkey.Values() {
				valStr := GetValueString(val)
				switch val.Name() {
				case "DisplayName":
					displayName = valStr
				case "DisplayVersion":
					displayVersion = valStr
				case "Publisher":
					publisher = valStr
				case "InstallDate":
					installDate = valStr
				case "UninstallString":
					uninstallString = valStr
				}
			}

			if displayName != "" {
				count++
				fmt.Printf("\n[%d] %s\n", count, displayName)
				if displayVersion != "" {
					fmt.Printf("    Version: %s\n", displayVersion)
				}
				if publisher != "" {
					fmt.Printf("    Publisher: %s\n", publisher)
				}
				if installDate != "" {
					fmt.Printf("    Install Date: %s\n", installDate)
				}
				if uninstallString != "" {
					fmt.Printf("    Uninstall: %s\n", uninstallString)
				}
				fmt.Printf("    Last Write: %s\n", subkey.Timestamp().Format("2006-01-02 15:04:05"))
			}
		}
	}

	fmt.Printf("\nTotal programs found: %d\n", count)
	return nil
}

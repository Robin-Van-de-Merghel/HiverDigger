package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&AppCompatPlugin{})
}

// AppCompatPlugin displays Application Compatibility flags and settings.
type AppCompatPlugin struct{}

func (p *AppCompatPlugin) Name() string {
	return "appcompat"
}

func (p *AppCompatPlugin) Description() string {
	return "Display Application Compatibility flags from NTUSER.DAT or SOFTWARE hive"
}

func (p *AppCompatPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT", "SOFTWARE"}
}

func (p *AppCompatPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Software/Microsoft/Windows NT/CurrentVersion/AppCompatFlags/Layers",
		"Software/Microsoft/Windows NT/CurrentVersion/AppCompatFlags/Compatibility Assistant/Store",
		"Microsoft/Windows NT/CurrentVersion/AppCompatFlags",
	}

	fmt.Println("Application Compatibility Settings:")
	fmt.Println(strings.Repeat("=", 80))

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("\n%s:\n", path)
		fmt.Printf("Last Write: %s\n\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() != "" {
				valStr := GetValueString(val)
				fmt.Printf("  %s\n", val.Name())
				if valStr != "" {
					fmt.Printf("    Flags: %s\n", valStr)
				}
			}
		}

		// Check subkeys
		for _, subkey := range key.Subkeys() {
			fmt.Printf("\n  Subkey: %s\n", subkey.Name())
			fmt.Printf("  Last Write: %s\n", subkey.Timestamp().Format("2006-01-02 15:04:05"))

			for _, val := range subkey.Values() {
				if val.Name() != "" {
					valStr := GetValueString(val)
					fmt.Printf("    %s = %s\n", val.Name(), valStr)
				}
			}
		}
	}

	if !found {
		fmt.Println("No Application Compatibility flags found")
	}

	return nil
}

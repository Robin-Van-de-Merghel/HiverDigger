package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&AutoRunPlugin{})
}

// AutoRunPlugin lists autorun locations from SOFTWARE hive.
type AutoRunPlugin struct{}

func (p *AutoRunPlugin) Name() string {
	return "autorun"
}

func (p *AutoRunPlugin) Description() string {
	return "List autorun locations from SOFTWARE hive"
}

func (p *AutoRunPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *AutoRunPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft\\Windows\\CurrentVersion\\Run",
		"Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"Microsoft\\Windows\\CurrentVersion\\Policies\\Explorer\\Run",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Run",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
	}

	fmt.Println("Autorun Locations")
	fmt.Println("=================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		if len(key.Values()) > 0 {
			fmt.Printf("[%s]\n", path)
			fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

			for _, val := range key.Values() {
				if val.Name() != "" {
					fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
				}
			}
			fmt.Println()
		}
	}

	return nil
}

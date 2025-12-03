package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ActiveSetupPlugin{})
}

// ActiveSetupPlugin displays Active Setup from SOFTWARE hive.
type ActiveSetupPlugin struct{}

func (p *ActiveSetupPlugin) Name() string {
	return "activesetup"
}

func (p *ActiveSetupPlugin) Description() string {
	return "Display Active Setup components from SOFTWARE hive"
}

func (p *ActiveSetupPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *ActiveSetupPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft\\Active Setup\\Installed Components",
		"Wow6432Node\\Microsoft\\Active Setup\\Installed Components",
	}

	fmt.Println("Active Setup Components")
	fmt.Println("=======================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)

		for _, component := range key.Subkeys() {
			stubPath := ""
			for _, val := range component.Values() {
				if val.Name() == "StubPath" {
					stubPath = GetValueString(val)
					break
				}
			}

			if stubPath != "" {
				fmt.Printf("  %s\n", component.Name())
				fmt.Printf("    StubPath: %s\n", stubPath)
				fmt.Printf("    Last Modified: %s\n", component.Timestamp().Format("2006-01-02 15:04:05"))
			}
		}
		fmt.Println()
	}

	return nil
}

package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&AppInitPlugin{})
}

// AppInitPlugin displays AppInit DLLs from SOFTWARE hive.
type AppInitPlugin struct{}

func (p *AppInitPlugin) Name() string {
	return "appinit"
}

func (p *AppInitPlugin) Description() string {
	return "Display AppInit DLLs from SOFTWARE hive"
}

func (p *AppInitPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *AppInitPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft\\Windows NT\\CurrentVersion\\Windows",
		"Wow6432Node\\Microsoft\\Windows NT\\CurrentVersion\\Windows",
	}

	fmt.Println("AppInit DLLs")
	fmt.Println("============")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() == "AppInit_DLLs" || val.Name() == "LoadAppInit_DLLs" {
				fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
			}
		}
		fmt.Println()
	}

	return nil
}

package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&AppPathsPlugin{})
}

// AppPathsPlugin displays App Paths from SOFTWARE hive.
type AppPathsPlugin struct{}

func (p *AppPathsPlugin) Name() string {
	return "apppaths"
}

func (p *AppPathsPlugin) Description() string {
	return "Display App Paths from SOFTWARE hive"
}

func (p *AppPathsPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *AppPathsPlugin) Run(hive *regf.Hive) error {
	path := "Microsoft\\Windows\\CurrentVersion\\App Paths"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("app paths key not found: %w", err)
	}

	fmt.Println("Application Paths")
	fmt.Println("=================")
	fmt.Println()

	for _, appKey := range key.Subkeys() {
		appName := appKey.Name()
		appPath := ""

		for _, val := range appKey.Values() {
			if val.Name() == "" || val.Name() == "(Default)" {
				appPath = GetValueString(val)
				break
			}
		}

		if appPath != "" {
			fmt.Printf("%s: %s\n", appName, appPath)
		}
	}

	return nil
}

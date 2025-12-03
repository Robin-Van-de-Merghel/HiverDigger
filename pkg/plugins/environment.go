package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&EnvironmentPlugin{})
}

// EnvironmentPlugin displays environment variables from SYSTEM hive.
type EnvironmentPlugin struct{}

func (p *EnvironmentPlugin) Name() string {
	return "environment"
}

func (p *EnvironmentPlugin) Description() string {
	return "Display environment variables from SYSTEM hive"
}

func (p *EnvironmentPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *EnvironmentPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Session Manager\\Environment",
		"ControlSet002\\Control\\Session Manager\\Environment",
	}

	fmt.Println("System Environment Variables")
	fmt.Println("============================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
		fmt.Println()

		for _, val := range key.Values() {
			if val.Name() != "" {
				fmt.Printf("%s = %s\n", val.Name(), GetValueString(val))
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("environment key not found")
}

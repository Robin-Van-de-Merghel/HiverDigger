package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&KnownDLLsPlugin{})
}

// KnownDLLsPlugin displays KnownDLLs from SYSTEM hive.
type KnownDLLsPlugin struct{}

func (p *KnownDLLsPlugin) Name() string {
	return "knowndlls"
}

func (p *KnownDLLsPlugin) Description() string {
	return "Display KnownDLLs from SYSTEM hive"
}

func (p *KnownDLLsPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *KnownDLLsPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Session Manager\\KnownDLLs",
		"ControlSet002\\Control\\Session Manager\\KnownDLLs",
	}

	fmt.Println("Known DLLs")
	fmt.Println("==========")
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
				fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("KnownDLLs key not found")
}

package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&BootExecutePlugin{})
}

// BootExecutePlugin displays boot execute commands from SYSTEM hive.
type BootExecutePlugin struct{}

func (p *BootExecutePlugin) Name() string {
	return "bootexecute"
}

func (p *BootExecutePlugin) Description() string {
	return "Display boot execute commands from SYSTEM hive"
}

func (p *BootExecutePlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *BootExecutePlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Session Manager",
		"ControlSet002\\Control\\Session Manager",
	}

	fmt.Println("Boot Execute Commands")
	fmt.Println("=====================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() == "BootExecute" {
				fmt.Printf("  %s\n", GetValueString(val))
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("session manager key not found")
}

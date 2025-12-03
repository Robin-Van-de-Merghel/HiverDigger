package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ComputerNamePlugin{})
}

// ComputerNamePlugin displays computer name from SYSTEM hive.
type ComputerNamePlugin struct{}

func (p *ComputerNamePlugin) Name() string {
	return "compname"
}

func (p *ComputerNamePlugin) Description() string {
	return "Display computer name from SYSTEM hive"
}

func (p *ComputerNamePlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *ComputerNamePlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\ComputerName\\ComputerName",
		"ControlSet002\\Control\\ComputerName\\ComputerName",
	}

	fmt.Println("Computer Name Information")
	fmt.Println("=========================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		for _, val := range key.Values() {
			if val.Name() == "ComputerName" {
				fmt.Printf("Computer Name: %s\n", GetValueString(val))
				fmt.Printf("Path: %s\n", path)
				fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
				fmt.Println()
				return nil
			}
		}
	}

	return fmt.Errorf("computer name not found")
}

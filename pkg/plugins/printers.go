package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&PrintersPlugin{})
}

// PrintersPlugin displays installed printers from SYSTEM hive.
type PrintersPlugin struct{}

func (p *PrintersPlugin) Name() string {
	return "printers"
}

func (p *PrintersPlugin) Description() string {
	return "Display installed printers from SYSTEM hive"
}

func (p *PrintersPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *PrintersPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Print\\Printers",
		"ControlSet002\\Control\\Print\\Printers",
	}

	fmt.Println("Installed Printers")
	fmt.Println("==================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		for _, printerKey := range key.Subkeys() {
			fmt.Printf("Printer: %s\n", printerKey.Name())

			for _, val := range printerKey.Values() {
				switch val.Name() {
				case "Port":
					fmt.Printf("  Port: %s\n", GetValueString(val))
				case "Print Processor":
					fmt.Printf("  Print Processor: %s\n", GetValueString(val))
				case "Printer Driver":
					fmt.Printf("  Driver: %s\n", GetValueString(val))
				}
			}

			fmt.Printf("  Last Modified: %s\n", printerKey.Timestamp().Format("2006-01-02 15:04:05"))
			fmt.Println()
		}

		return nil
	}

	return fmt.Errorf("printers key not found")
}

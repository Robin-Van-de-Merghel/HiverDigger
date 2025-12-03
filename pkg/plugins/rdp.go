package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&TerminalServerPlugin{})
}

// TerminalServerPlugin displays Terminal Server/RDP information from SYSTEM hive.
type TerminalServerPlugin struct{}

func (p *TerminalServerPlugin) Name() string {
	return "rdp"
}

func (p *TerminalServerPlugin) Description() string {
	return "Display Terminal Server/RDP information from SYSTEM hive"
}

func (p *TerminalServerPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *TerminalServerPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Terminal Server",
		"ControlSet002\\Control\\Terminal Server",
	}

	fmt.Println("Terminal Server/RDP Configuration")
	fmt.Println("==================================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("Terminal Server key not found")
}

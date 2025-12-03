package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&SessionManagerPlugin{})
}

// SessionManagerPlugin displays Session Manager information from SYSTEM hive.
type SessionManagerPlugin struct{}

func (p *SessionManagerPlugin) Name() string {
	return "sessionmgr"
}

func (p *SessionManagerPlugin) Description() string {
	return "Display Session Manager information from SYSTEM hive"
}

func (p *SessionManagerPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *SessionManagerPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Session Manager",
		"ControlSet002\\Control\\Session Manager",
	}

	fmt.Println("Session Manager Configuration")
	fmt.Println("=============================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() == "BootExecute" || val.Name() == "PendingFileRenameOperations" {
				fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("Session Manager key not found")
}

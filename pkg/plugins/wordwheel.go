package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&WordWheelQueryPlugin{})
}

// WordWheelQueryPlugin displays Windows search terms from NTUSER.DAT.
type WordWheelQueryPlugin struct{}

func (p *WordWheelQueryPlugin) Name() string {
	return "wordwheel"
}

func (p *WordWheelQueryPlugin) Description() string {
	return "Display Windows search terms from NTUSER.DAT"
}

func (p *WordWheelQueryPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *WordWheelQueryPlugin) Run(hive *regf.Hive) error {
	path := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\WordWheelQuery"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("WordWheelQuery key not found: %w", err)
	}

	fmt.Println("Windows Search Terms")
	fmt.Println("====================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	for _, val := range key.Values() {
		if val.Name() == "MRUListEx" {
			// Binary data representing order
			continue
		}
		if val.Name() != "" {
			fmt.Printf("%s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}

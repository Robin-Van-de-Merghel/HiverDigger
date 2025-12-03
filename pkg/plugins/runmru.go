package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&RunMRUPlugin{})
}

// RunMRUPlugin displays Run dialog MRU from NTUSER.DAT.
type RunMRUPlugin struct{}

func (p *RunMRUPlugin) Name() string {
	return "runmru"
}

func (p *RunMRUPlugin) Description() string {
	return "Display Run dialog history from NTUSER.DAT"
}

func (p *RunMRUPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *RunMRUPlugin) Run(hive *regf.Hive) error {
	path := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\RunMRU"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("RunMRU key not found: %w", err)
	}

	fmt.Println("Run Dialog History (MRU)")
	fmt.Println("========================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Get MRU order
	var mruOrder string
	for _, val := range key.Values() {
		if val.Name() == "MRUList" {
			mruOrder = GetValueString(val)
			break
		}
	}

	// Display in MRU order
	if mruOrder != "" {
		for i, char := range mruOrder {
			valName := string(char)
			for _, val := range key.Values() {
				if val.Name() == valName {
					fmt.Printf("%d. %s\n", i+1, GetValueString(val))
					break
				}
			}
		}
	} else {
		// No MRU order, just list all
		for _, val := range key.Values() {
			if val.Name() != "" && val.Name() != "MRUList" {
				fmt.Printf("%s: %s\n", val.Name(), GetValueString(val))
			}
		}
	}

	return nil
}

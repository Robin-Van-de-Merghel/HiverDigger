package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&MUICachePlugin{})
}

// MUICachePlugin displays MUICache entries showing executed applications.
// Based on RegRipper's muicache.pl plugin.
type MUICachePlugin struct{}

func (p *MUICachePlugin) Name() string {
	return "muicache"
}

func (p *MUICachePlugin) Description() string {
	return "Display MUICache entries (executed applications) from USRCLASS.DAT hive"
}

func (p *MUICachePlugin) CompatibleHiveTypes() []string {
	return []string{"USRCLASS.DAT", "NTUSER.DAT"}
}

func (p *MUICachePlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Local Settings/Software/Microsoft/Windows/Shell/MuiCache",
		"Software/Microsoft/Windows/ShellNoRoam/MUICache",
		"Software/Classes/Local Settings/Software/Microsoft/Windows/Shell/MuiCache",
	}

	fmt.Println("MUICache Entries (Executed Applications):")
	fmt.Println(strings.Repeat("=", 80))

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("\n%s:\n", path)
		fmt.Printf("Last Write Time: %s\n\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() != "" {
				valStr := GetValueString(val)
				fmt.Printf("  %s\n", val.Name())
				if valStr != "" {
					fmt.Printf("    Description: %s\n", valStr)
				}
			}
		}
	}

	if !found {
		fmt.Println("No MUICache entries found")
	}

	return nil
}

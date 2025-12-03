package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&JumpListsPlugin{})
}

// JumpListsPlugin extracts Jump List artifacts from user registry
type JumpListsPlugin struct{}

func (p *JumpListsPlugin) Name() string {
	return "jumplists"
}

func (p *JumpListsPlugin) Description() string {
	return "Jump List artifacts - recent and frequent items"
}

func (p *JumpListsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT", "USRCLASS.DAT"}
}

func (p *JumpListsPlugin) Run(hive *regf.Hive) error {
	// Try TaskBand (Win7+)
	taskbandPath := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\TaskBand"
	if taskband, err := hive.GetKey(taskbandPath); err == nil {
		fmt.Println("TaskBand Jump Lists:")
		fmt.Println(strings.Repeat("=", 80))
		
		for _, val := range taskband.Values() {
			if strings.Contains(val.Name(), "Favorites") {
				fmt.Printf("%s: %d bytes\n", val.Name(), len(val.Bytes()))
			}
		}
	}

	// Try Destinations (Win7+)
	destPath := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\FeatureUsage\\AppSwitched"
	if dest, err := hive.GetKey(destPath); err == nil {
		fmt.Println("\nRecently Used Applications:")
		fmt.Println(strings.Repeat("=", 80))
		
		for _, val := range dest.Values() {
			fmt.Printf("%s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}

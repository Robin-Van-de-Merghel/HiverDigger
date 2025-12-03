package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&WindowsVersionPlugin{})
}

// WindowsVersionPlugin displays Windows version from SOFTWARE hive.
type WindowsVersionPlugin struct{}

func (p *WindowsVersionPlugin) Name() string {
	return "winver"
}

func (p *WindowsVersionPlugin) Description() string {
	return "Display Windows version from SOFTWARE hive"
}

func (p *WindowsVersionPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *WindowsVersionPlugin) Run(hive *regf.Hive) error {
	path := "Microsoft\\Windows NT\\CurrentVersion"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("Windows version key not found: %w", err)
	}

	fmt.Println("Windows Version Information")
	fmt.Println("===========================")
	fmt.Println()

	values := []string{
		"ProductName",
		"CurrentVersion",
		"CurrentBuild",
		"CurrentBuildNumber",
		"InstallDate",
		"RegisteredOwner",
		"RegisteredOrganization",
		"ProductId",
		"EditionID",
		"CompositionEditionID",
	}

	for _, valName := range values {
		for _, val := range key.Values() {
			if val.Name() == valName {
				fmt.Printf("%s: %s\n", valName, GetValueString(val))
				break
			}
		}
	}

	fmt.Printf("\nLast Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

	return nil
}

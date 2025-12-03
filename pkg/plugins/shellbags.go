package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ShellBagsPlugin{})
}

// ShellBagsPlugin displays ShellBags data from NTUSER.DAT and UsrClass.dat.
type ShellBagsPlugin struct{}

func (p *ShellBagsPlugin) Name() string {
	return "shellbags"
}

func (p *ShellBagsPlugin) Description() string {
	return "Display ShellBags data (folder access history)"
}

func (p *ShellBagsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT", "USRCLASS.DAT"}
}

func (p *ShellBagsPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Software\\Microsoft\\Windows\\Shell\\Bags",
		"Software\\Microsoft\\Windows\\Shell\\BagMRU",
		"Software\\Microsoft\\Windows\\ShellNoRoam\\Bags",
		"Software\\Microsoft\\Windows\\ShellNoRoam\\BagMRU",
		"Software\\Classes\\Local Settings\\Software\\Microsoft\\Windows\\Shell\\Bags",
		"Software\\Classes\\Local Settings\\Software\\Microsoft\\Windows\\Shell\\BagMRU",
	}

	fmt.Println("ShellBags (Folder Access History)")
	fmt.Println("==================================")
	fmt.Println()

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
		fmt.Printf("Subkeys: %d\n", len(key.Subkeys()))
		fmt.Println()
	}

	if !found {
		return fmt.Errorf("no ShellBags keys found")
	}

	return nil
}

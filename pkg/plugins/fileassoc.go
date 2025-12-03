package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&FileAssocPlugin{})
}

// FileAssocPlugin displays file associations from SOFTWARE hive.
type FileAssocPlugin struct{}

func (p *FileAssocPlugin) Name() string {
	return "fileassoc"
}

func (p *FileAssocPlugin) Description() string {
	return "Display file associations from SOFTWARE hive"
}

func (p *FileAssocPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *FileAssocPlugin) Run(hive *regf.Hive) error {
	path := "Classes"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("classes key not found: %w", err)
	}

	fmt.Println("File Associations")
	fmt.Println("=================")
	fmt.Println()

	count := 0
	for _, subkey := range key.Subkeys() {
		name := subkey.Name()
		if len(name) > 0 && name[0] == '.' {
			// This is a file extension
			defaultValue := ""
			for _, val := range subkey.Values() {
				if val.Name() == "" || val.Name() == "(Default)" {
					defaultValue = GetValueString(val)
					break
				}
			}

			if defaultValue != "" {
				fmt.Printf("%s -> %s\n", name, defaultValue)
				count++
				if count >= 50 {
					fmt.Println("... (truncated)")
					break
				}
			}
		}
	}

	return nil
}

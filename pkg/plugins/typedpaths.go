package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&TypedPathsPlugin{})
}

// TypedPathsPlugin displays typed paths from NTUSER.DAT.
type TypedPathsPlugin struct{}

func (p *TypedPathsPlugin) Name() string {
	return "typedpaths"
}

func (p *TypedPathsPlugin) Description() string {
	return "Display typed paths from Windows Explorer (NTUSER.DAT)"
}

func (p *TypedPathsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *TypedPathsPlugin) Run(hive *regf.Hive) error {
	path := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\TypedPaths"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("TypedPaths key not found: %w", err)
	}

	fmt.Println("Typed Paths (Windows Explorer)")
	fmt.Println("==============================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	for _, val := range key.Values() {
		if val.Name() != "" {
			fmt.Printf("%s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}

package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&TypedURLsPlugin{})
}

// TypedURLsPlugin displays typed URLs from NTUSER.DAT.
type TypedURLsPlugin struct{}

func (p *TypedURLsPlugin) Name() string {
	return "typedurls"
}

func (p *TypedURLsPlugin) Description() string {
	return "Display typed URLs from Internet Explorer (NTUSER.DAT)"
}

func (p *TypedURLsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *TypedURLsPlugin) Run(hive *regf.Hive) error {
	path := "Software\\Microsoft\\Internet Explorer\\TypedURLs"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("TypedURLs key not found: %w", err)
	}

	fmt.Println("Typed URLs (Internet Explorer)")
	fmt.Println("==============================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	for _, val := range key.Values() {
		if val.Name() != "" && val.Name() != "(Default)" {
			fmt.Printf("%s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}

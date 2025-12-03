package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&BrowserHelperPlugin{})
}

// BrowserHelperPlugin displays Browser Helper Objects from SOFTWARE hive.
type BrowserHelperPlugin struct{}

func (p *BrowserHelperPlugin) Name() string {
	return "bho"
}

func (p *BrowserHelperPlugin) Description() string {
	return "Display Browser Helper Objects from SOFTWARE hive"
}

func (p *BrowserHelperPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *BrowserHelperPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft\\Windows\\CurrentVersion\\Explorer\\Browser Helper Objects",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Explorer\\Browser Helper Objects",
	}

	fmt.Println("Browser Helper Objects (BHO)")
	fmt.Println("============================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)

		for _, bhoKey := range key.Subkeys() {
			fmt.Printf("  CLSID: %s\n", bhoKey.Name())

			// Try to get the name
			for _, val := range bhoKey.Values() {
				if val.Name() == "" || val.Name() == "(Default)" {
					name := GetValueString(val)
					if name != "" {
						fmt.Printf("    Name: %s\n", name)
					}
				}
			}

			fmt.Printf("    Last Modified: %s\n", bhoKey.Timestamp().Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	return nil
}

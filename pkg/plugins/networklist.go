package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&NetworkListPlugin{})
}

// NetworkListPlugin lists network profiles from SOFTWARE hive.
type NetworkListPlugin struct{}

func (p *NetworkListPlugin) Name() string {
	return "networklist"
}

func (p *NetworkListPlugin) Description() string {
	return "List network profiles from SOFTWARE hive"
}

func (p *NetworkListPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *NetworkListPlugin) Run(hive *regf.Hive) error {
	path := "Microsoft\\Windows NT\\CurrentVersion\\NetworkList\\Profiles"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("network list key not found: %w", err)
	}

	fmt.Println("Network Profiles")
	fmt.Println("================")
	fmt.Println()

	for _, profile := range key.Subkeys() {
		fmt.Printf("Profile GUID: %s\n", profile.Name())

		for _, val := range profile.Values() {
			name := val.Name()
			switch {
			case strings.EqualFold(name, "ProfileName"):
				fmt.Printf("  Name: %s\n", GetValueString(val))
			case strings.EqualFold(name, "Description"):
				fmt.Printf("  Description: %s\n", GetValueString(val))
			case strings.EqualFold(name, "DateCreated"):
				// Could parse this as a binary timestamp
				fmt.Printf("  Date Created: (binary data)\n")
			case strings.EqualFold(name, "DateLastConnected"):
				fmt.Printf("  Date Last Connected: (binary data)\n")
			}
		}

		fmt.Printf("  Last Modified: %s\n", profile.Timestamp().Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

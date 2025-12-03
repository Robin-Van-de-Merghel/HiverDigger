package plugins

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&SAMUsersPlugin{})
}

// SAMUsersPlugin lists local users from SAM hive.
type SAMUsersPlugin struct{}

func (p *SAMUsersPlugin) Name() string {
	return "samusers"
}

func (p *SAMUsersPlugin) Description() string {
	return "List local users from SAM hive"
}

func (p *SAMUsersPlugin) CompatibleHiveTypes() []string {
	return []string{"SAM"}
}

func (p *SAMUsersPlugin) Run(hive *regf.Hive) error {
	path := "SAM\\Domains\\Account\\Users"

	_, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("SAM Users key not found: %w", err)
	}

	fmt.Println("Local Users (from SAM)")
	fmt.Println("======================")
	fmt.Println()

	// Get Names subkey for username mapping
	namesKey, err := hive.GetKey(path + "\\Names")
	if err == nil {
		for _, nameKey := range namesKey.Subkeys() {
			username := nameKey.Name()

			// Get RID from the default value type field
			rid := ""
			for _, val := range nameKey.Values() {
				if val.Name() == "" || strings.EqualFold(val.Name(), "(Default)") {
					data := val.Bytes()
					if len(data) >= 4 {
						ridNum := binary.LittleEndian.Uint32(data)
						rid = fmt.Sprintf("0x%x", ridNum)
					}
					break
				}
			}

			fmt.Printf("Username: %s\n", username)
			if rid != "" {
				fmt.Printf("  RID: %s\n", rid)
			}
			fmt.Printf("  Last Modified: %s\n", nameKey.Timestamp().Format("2006-01-02 15:04:05"))
			fmt.Println()
		}
	}

	return nil
}

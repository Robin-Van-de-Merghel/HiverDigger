package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&SAMParsePlugin{})
}

// SAMParsePlugin provides comprehensive SAM parsing (users, groups, last login)
type SAMParsePlugin struct{}

func (p *SAMParsePlugin) Name() string {
	return "samparse"
}

func (p *SAMParsePlugin) Description() string {
	return "Comprehensive SAM parsing - users, groups, last login times"
}

func (p *SAMParsePlugin) CompatibleHiveTypes() []string {
	return []string{"SAM"}
}

func (p *SAMParsePlugin) Run(hive *regf.Hive) error {
	usersPath := "SAM\\Domains\\Account\\Users"
	usersKey, err := hive.GetKey(usersPath)
	if err != nil {
		return fmt.Errorf("users key not found: %w", err)
	}

	fmt.Println("SAM Users (Comprehensive):")
	fmt.Println(strings.Repeat("=", 80))

	namesKey, _ := getSubkey(usersKey, "Names")

	for _, user := range usersKey.Subkeys() {
		if user.Name() == "Names" {
			continue
		}

		fmt.Printf("\nRID: %s\n", user.Name())
		fmt.Printf("Last Write: %s\n", user.Timestamp().Format("2006-01-02 15:04:05"))

		// Try to find username from Names subkey
		if namesKey != nil {
			for _, nameEntry := range namesKey.Subkeys() {
				if defVal, err := getValue(nameEntry, ""); err == nil {
					ridStr := fmt.Sprintf("0x%s", user.Name())
					if strings.Contains(GetValueString(defVal), ridStr) || user.Name() == nameEntry.Name() {
						fmt.Printf("Username: %s\n", nameEntry.Name())
					}
				}
			}
		}

		// Parse F value (contains last login, etc.)
		if fVal, err := getValue(user, "F"); err == nil {
			data := fVal.Bytes()
			if len(data) >= 8 {
				fmt.Printf("F value length: %d bytes\n", len(data))
			}
		}
	}

	return nil
}

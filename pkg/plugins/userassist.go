package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&UserAssistPlugin{})
}

// UserAssistPlugin decodes and displays UserAssist data from NTUSER.DAT.
type UserAssistPlugin struct{}

func (p *UserAssistPlugin) Name() string {
	return "userassist"
}

func (p *UserAssistPlugin) Description() string {
	return "Display UserAssist data from NTUSER.DAT (program execution)"
}

func (p *UserAssistPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT", "USRCLASS.DAT"}
}

func (p *UserAssistPlugin) Run(hive *regf.Hive) error {
	basePath := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\UserAssist"

	key, err := hive.GetKey(basePath)
	if err != nil {
		return fmt.Errorf("UserAssist key not found: %w", err)
	}

	fmt.Println("UserAssist Data (Program Execution)")
	fmt.Println("====================================")
	fmt.Println()

	for _, guidKey := range key.Subkeys() {
		fmt.Printf("GUID: %s\n", guidKey.Name())

		countKey, err := hive.GetKey(basePath + "\\" + guidKey.Name() + "\\Count")
		if err != nil {
			continue
		}

		for _, val := range countKey.Values() {
			if val.Name() != "" {
				// ROT13 decode the name
				decodedName := rot13(val.Name())
				fmt.Printf("  %s\n", decodedName)
			}
		}
		fmt.Println()
	}

	return nil
}

// rot13 decodes ROT13 encoded strings used in UserAssist
func rot13(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = 'A' + (c-'A'+13)%26
		} else if c >= 'a' && c <= 'z' {
			result[i] = 'a' + (c-'a'+13)%26
		} else {
			result[i] = c
		}
	}
	return string(result)
}

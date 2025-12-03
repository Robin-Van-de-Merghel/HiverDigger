package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&WinlogonPlugin{})
}

// WinlogonPlugin displays Winlogon information from SOFTWARE hive.
type WinlogonPlugin struct{}

func (p *WinlogonPlugin) Name() string {
	return "winlogon"
}

func (p *WinlogonPlugin) Description() string {
	return "Display Winlogon information from SOFTWARE hive"
}

func (p *WinlogonPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *WinlogonPlugin) Run(hive *regf.Hive) error {
	path := "Microsoft\\Windows NT\\CurrentVersion\\Winlogon"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("winlogon key not found: %w", err)
	}

	fmt.Println("Winlogon Information")
	fmt.Println("====================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	interestingValues := []string{
		"DefaultUserName",
		"DefaultDomainName",
		"AutoAdminLogon",
		"Shell",
		"Userinit",
		"VmApplet",
		"LegalNoticeCaption",
		"LegalNoticeText",
	}

	for _, valName := range interestingValues {
		for _, val := range key.Values() {
			if val.Name() == valName {
				fmt.Printf("%s: %s\n", valName, GetValueString(val))
				break
			}
		}
	}

	return nil
}
